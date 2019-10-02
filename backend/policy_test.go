package main

import (
	"github.com/casbin/casbin"
	"github.com/casbin/casbin/util"
	"testing"
)


func enforcePolicy(params ...interface{}) bool {
	// Ready the model and the policy from the files. In the real case
	// they will be read from the database.
	e := casbin.NewEnforcer("./model.conf", "./policy.csv")

	// There are several standard matching functions used in the models
	// but some of them are not registered by default. The keyMatch3 is
	// one of them.
	e.AddFunction("keyMatch3", util.KeyMatch3Func)

	res, err := e.EnforceSafe(params...)

	if err != nil {
		panic(err.Error())
	}

	return res
}

func TestPolicies(t *testing.T) {
	// User marcin can manage server 2
	if !enforcePolicy("marcin", "/servers/2/subnets/3", "POST") {
		t.Errorf("user marcin should be able to manage server 2")
	}

	// User xiong can't manage server 2
	if enforcePolicy("xiong", "/servers/2/subnets/3", "POST") {
		t.Errorf("user xiong should not be able to manage server 2")
	}

	// but user xiong can view server 3
	if !enforcePolicy("xiong", "/servers/3/subnets/3", "GET") {
		t.Errorf("user xiong should be able to view server 3")
	}

	// user xiong cannot modify server 3
	if enforcePolicy("xiong", "/servers/3/subnets/3", "POST") {
		t.Errorf("user xiong should not be able to modify server 3")
	}

	// user marcin should be able to list all subnets
	if !enforcePolicy("marcin", "/subnets", "GET") {
		t.Errorf("user marcin should be able to view fetched subnets")
	}

	// but user xiong shouldn't be able to list all subnets
	if enforcePolicy("xiong", "/subnets", "GET") {
		t.Errorf("user xiong should not be able to view fetched subnets")
	}

	// marcin should be able to manage a machine as he belongs to the
	// machine_managers
	if !enforcePolicy("marcin", "/machines/1/os", "POST") {
		t.Errorf("user marcin should be able to manage the machine elements")
	}
}
