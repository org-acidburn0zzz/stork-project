package agent

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"isc.org/stork"
	agentapi "isc.org/stork/api"
)

// Stork Agent settings.
type Settings struct {
	Host string `long:"host" description:"the IP to listen on" env:"STORK_AGENT_ADDRESS"`
	Port int    `long:"port" description:"the port to listen on for connections" default:"8080" env:"STORK_AGENT_PORT"`
}

// Global Stork Agent state
type StorkAgent struct {
	Settings   Settings
	AppMonitor AppMonitor

	HTTPClient *HTTPClient // to communicate with Kea Control Agent and named statistics-channel
	RndcClient *RndcClient // to communicate with BIND 9 via rndc
	server     *grpc.Server
}

// API exposed to Stork Server

func NewStorkAgent(appMonitor AppMonitor) *StorkAgent {
	// rndc is the command to interface with BIND 9.
	rndc := func(command []string) ([]byte, error) {
		cmd := exec.Command(command[0], command[1:]...) //nolint:gosec
		return cmd.Output()
	}
	rndcClient := NewRndcClient(rndc)

	httpClient := NewHTTPClient()

	server := grpc.NewServer()

	sa := &StorkAgent{
		AppMonitor: appMonitor,
		HTTPClient: httpClient,
		RndcClient: rndcClient,
		server:     server,
	}

	return sa
}

// Get state of machine.
func (sa *StorkAgent) GetState(ctx context.Context, in *agentapi.GetStateReq) (*agentapi.GetStateRsp, error) {
	vm, _ := mem.VirtualMemory()
	hostInfo, _ := host.Info()
	load, _ := load.Avg()
	loadStr := fmt.Sprintf("%.2f %.2f %.2f", load.Load1, load.Load5, load.Load15)

	var apps []*agentapi.App
	for _, app := range sa.AppMonitor.GetApps() {
		var accessPoints []*agentapi.AccessPoint
		for _, point := range app.AccessPoints {
			accessPoints = append(accessPoints, &agentapi.AccessPoint{
				Type:    point.Type,
				Address: point.Address,
				Port:    point.Port,
				Key:     point.Key,
			})
		}

		apps = append(apps, &agentapi.App{
			Type:         app.Type,
			AccessPoints: accessPoints,
		})
	}

	state := agentapi.GetStateRsp{
		AgentVersion:         stork.Version,
		Apps:                 apps,
		Hostname:             hostInfo.Hostname,
		Cpus:                 int64(runtime.NumCPU()),
		CpusLoad:             loadStr,
		Memory:               int64(vm.Total / (1024 * 1024 * 1024)), // in GiB
		UsedMemory:           int64(vm.UsedPercent),
		Uptime:               int64(hostInfo.Uptime / (60 * 60 * 24)), // in days
		Os:                   hostInfo.OS,
		Platform:             hostInfo.Platform,
		PlatformFamily:       hostInfo.PlatformFamily,
		PlatformVersion:      hostInfo.PlatformVersion,
		KernelVersion:        hostInfo.KernelVersion,
		KernelArch:           hostInfo.KernelArch,
		VirtualizationSystem: hostInfo.VirtualizationSystem,
		VirtualizationRole:   hostInfo.VirtualizationRole,
		HostID:               hostInfo.HostID,
		Error:                "",
	}

	return &state, nil
}

// ForwardRndcCommand forwards one rndc command sent by the Stork server to
// the named daemon.
func (sa *StorkAgent) ForwardRndcCommand(ctx context.Context, in *agentapi.ForwardRndcCommandReq) (*agentapi.ForwardRndcCommandRsp, error) {
	accessPoints := []AccessPoint{
		{
			Type:    AccessPointControl,
			Address: in.Address,
			Port:    in.Port,
			Key:     in.Key,
		},
	}

	app := &App{
		Type:         AppTypeBind9,
		AccessPoints: accessPoints,
	}

	request := in.GetRndcRequest()
	response := &agentapi.ForwardRndcCommandRsp{
		Status: &agentapi.Status{
			Code: agentapi.Status_OK, // all ok
		},
	}

	rndcRsp := &agentapi.RndcResponse{
		Status: &agentapi.Status{},
	}

	// Try to forward the command to rndc.
	output, err := sa.RndcClient.Call(app, strings.Fields(request.Request))
	if err != nil {
		log.WithFields(log.Fields{
			"Address": accessPoints[0].Address,
			"Port":    accessPoints[0].Port,
			"Key":     accessPoints[0].Key,
		}).Errorf("Failed to forward commands to rndc: %+v", err)
		rndcRsp.Status.Code = agentapi.Status_ERROR
		rndcRsp.Status.Message = "Failed to forward commands to rndc"
	} else {
		rndcRsp.Status.Code = agentapi.Status_OK
		rndcRsp.Response = string(output)
	}

	response.Status = rndcRsp.Status
	response.RndcResponse = rndcRsp
	return response, err
}

// ForwardToNamedStats forwards a statistics request to the named daemon.
func (sa *StorkAgent) ForwardToNamedStats(ctx context.Context, in *agentapi.ForwardToNamedStatsReq) (*agentapi.ForwardToNamedStatsRsp, error) {
	reqURL := in.GetUrl()
	req := in.GetNamedStatsRequest()

	response := &agentapi.ForwardToNamedStatsRsp{
		Status: &agentapi.Status{
			Code: agentapi.Status_OK, // all ok
		},
	}

	rsp := &agentapi.NamedStatsResponse{
		Status: &agentapi.Status{},
	}
	// Try to forward the command to named daemon.
	namedRsp, err := sa.HTTPClient.Call(reqURL, bytes.NewBuffer([]byte(req.Request)))
	if err != nil {
		log.WithFields(log.Fields{
			"URL": reqURL,
		}).Errorf("Failed to forward commands to named: %+v", err)
		rsp.Status.Code = agentapi.Status_ERROR
		rsp.Status.Message = "Failed to forward commands to named"
		response.NamedStatsResponse = rsp
		return response, nil
	}

	// Read the response body.
	body, err := ioutil.ReadAll(namedRsp.Body)
	namedRsp.Body.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"URL": reqURL,
		}).Errorf("Failed to read the body of the named response: %+v", err)
		rsp.Status.Code = agentapi.Status_ERROR
		rsp.Status.Message = "Failed to read the body of the named response"
		response.NamedStatsResponse = rsp
		return response, nil
	}

	// Everything looks good, so include the body in the response.
	rsp.Response = string(body)
	rsp.Status.Code = agentapi.Status_OK
	response.NamedStatsResponse = rsp
	return response, nil
}

// Forwards one or more Kea commands sent by the Stork server to the appropriate Kea instance over
// HTTP (via Control Agent).
func (sa *StorkAgent) ForwardToKeaOverHTTP(ctx context.Context, in *agentapi.ForwardToKeaOverHTTPReq) (*agentapi.ForwardToKeaOverHTTPRsp, error) {
	reqURL := in.GetUrl()

	requests := in.GetKeaRequests()

	response := &agentapi.ForwardToKeaOverHTTPRsp{
		Status: &agentapi.Status{
			Code: agentapi.Status_OK, // all ok
		},
	}

	// forward requests to kea one by one
	for _, req := range requests {
		rsp := &agentapi.KeaResponse{
			Status: &agentapi.Status{},
		}
		// Try to forward the command to Kea Control Agent.
		keaRsp, err := sa.HTTPClient.Call(reqURL, bytes.NewBuffer([]byte(req.Request)))
		if err != nil {
			log.WithFields(log.Fields{
				"URL": reqURL,
			}).Errorf("Failed to forward commands to Kea: %+v", err)
			rsp.Status.Code = agentapi.Status_ERROR
			rsp.Status.Message = "Failed to forward commands to Kea"
			response.KeaResponses = append(response.KeaResponses, rsp)
			continue
		}

		// Read the response body.
		body, err := ioutil.ReadAll(keaRsp.Body)
		keaRsp.Body.Close()
		if err != nil {
			log.WithFields(log.Fields{
				"URL": reqURL,
			}).Errorf("Failed to read the body of the Kea response to forwarded commands: %+v", err)
			rsp.Status.Code = agentapi.Status_ERROR
			rsp.Status.Message = "Failed to read the body of the Kea response"
			response.KeaResponses = append(response.KeaResponses, rsp)
			continue
		}

		// Everything looks good, so include the body in the response.
		rsp.Response = string(body)
		rsp.Status.Code = agentapi.Status_OK
		response.KeaResponses = append(response.KeaResponses, rsp)
	}

	return response, nil
}

func (sa *StorkAgent) Serve() {
	// Install gRPC API handlers.
	agentapi.RegisterAgentServer(sa.server, sa)

	// Prepare listener on configured address.
	addr := fmt.Sprintf("%s:%d", sa.Settings.Host, sa.Settings.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on port: %+v", err)
	}

	// Start serving gRPC
	log.WithFields(log.Fields{
		"address": lis.Addr(),
	}).Infof("started serving Stork Agent")
	if err := sa.server.Serve(lis); err != nil {
		log.Fatalf("Failed to listen on port: %+v", err)
	}
}

func (sa *StorkAgent) Shutdown() {
	log.Infof("stopping StorkAgent")
	if sa.server != nil {
		sa.server.GracefulStop()
	}
}
