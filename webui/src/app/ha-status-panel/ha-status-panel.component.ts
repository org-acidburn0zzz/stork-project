import { Component, Input, OnInit } from '@angular/core'
import { datetimeToLocal } from '../utils'

/**
 * Component presenting live status of High Availability in Kea for
 * a single server.
 *
 * This component is embedded within the HaStatusComponent to present
 * the status of the individual servers.
 */
@Component({
    selector: 'app-ha-status-panel',
    templateUrl: './ha-status-panel.component.html',
    styleUrls: ['./ha-status-panel.component.sass'],
})
export class HaStatusPanelComponent implements OnInit {
    /**
     * Holds status fetched for this server from the backend.
     */
    private _serverStatus

    /**
     * Holds the style class of the panel view.
     *
     * The code may dynamically switch to a different class depending
     * on the server status. Switching to a different class causes the
     * panel to change its color to highlight warnings and errors.
     */
    public statusPanelClass = 'green-colored-panel'

    /**
     * Panel title set by the parent component.
     */
    @Input()
    public panelTitle: string

    /**
     * Server name set by the parent component.
     */
    @Input()
    public serverName: string

    /**
     * Indicates if monitoring one or two active servers.
     *
     * Two active servers are monitored in load-balancing and hot-standby
     * modes. A single server is monitored in the passive-backup mode. The
     * panel view is adjusted if this is single server case, e.g. the
     * last failover time is not presented.
     */
    @Input()
    public singleActiveServer = false

    /**
     * Indicates if the link to the app should be presented in the panel title.
     *
     * The link is presented for the remote servers. It is not presented for
     * the local servers. The parent component sets this flag.
     */
    @Input()
    public showServerLink = false

    /**
     * No-op constructor.
     */
    constructor() {}

    /**
     * No-op initialization.
     *
     * The pre-initialization is performed by the parent component.
     */
    ngOnInit(): void {}

    /**
     * Sets new status information for the server.
     *
     * The panel colors are refreshed according to the new status
     * information to highlight errors, warnings or normal operation.
     *
     * @param serverStatus New server status.
     */
    @Input()
    set serverStatus(serverStatus) {
        this._serverStatus = serverStatus
        this.refreshPanelColors()
    }

    /**
     * Returns status information fetched for the server.
     */
    get serverStatus() {
        return this._serverStatus
    }

    /**
     * Checks if the parent component has fetched the HA status for the server.
     *
     * @returns true if the status has been fetched and is available for display.
     */
    private hasStatus(): boolean {
        return this._serverStatus
    }

    /**
     * Returns the help tip describing online/offline control status.
     */
    controlStatusHelptip(): string {
        if (!this.hasStatus()) {
            return ''
        }
        if (this.serverStatus.inTouch) {
            return 'Server responds to the commands over the control channel.'
        }
        return 'Server does not respond to the commands over the control channel. It may be down!'
    }

    /**
     * Returns help tip describing various HA states.
     */
    haStateHelptip(): string {
        if (!this.hasStatus()) {
            return ''
        }
        switch (this.serverStatus.state) {
            case 'load-balancing':
            case 'hot-standby':
                return 'Normal operation.'
            case 'partner-down':
                return (
                    'This server now responds to all DHCP queries because it detected ' +
                    'that partner server is not functional!'
                )
            case 'passive-backup':
                return (
                    'The server has no active partner like in load-balancing or hot-standby ' +
                    'mode. This server may be configured to send lease updates to the ' +
                    'backup servers, but there is no automatic failover triggered in case ' +
                    'of failure.'
                )
            case 'waiting':
                return 'This server is apparently booting up and will try to synchronize its lease database.'
            case 'syncing':
                return 'This server is synchronizing its database after failure.'
            case 'ready':
                return 'This server synchronized its lease database and will start normal operation shortly.'
            case 'terminated':
                return 'This server no longer participates in the HA setup because of the too high clock skew.'
            case 'maintained':
                return 'This server is under maintenance.'
            case 'partner-maintained':
                return 'This server responds to all DHCP queries for the partner being in maintenance.'
            case 'unavailable':
                return 'Communication with the server failed. It may have crashed or have been shut down.'
            default:
                return 'Refer to Kea manual for details.'
        }
        return ''
    }

    /**
     * Returns help tip for last failover time.
     */
    failoverHelptip(): string {
        return (
            'This is the last time when the ' +
            this.serverName +
            ' server went to the partner-down state ' +
            'because its partner was considered offline as a result of unexpected termination ' +
            'or shutdown.'
        )
    }

    /**
     * Returns help tip for status time.
     */
    statusTimeHelptip(): string {
        return (
            'This is the time when the ' +
            this.serverName +
            ' server reported its state for the last time. ' +
            'This is not neccessarily the time when the state information ' +
            'was refreshed in the UI. The presented state information is ' +
            'typically delayed by 10 to 30 seconds because it is cached by the Kea ' +
            'servers and the Stork backend. Caching minimizes the performance ' +
            'impact on the DHCP servers reporting their states over the control ' +
            'channels.'
        )
    }

    /**
     * Returns help tip for status age.
     *
     * The age indicates how long ago the given server reported its status.
     */
    collectedHelptip(): string {
        return (
            'This is the duration between the "Status Time" and now, i.e. informs ' +
            'how long ago the ' +
            this.serverName +
            ' server reported its state. The long duration ' +
            'indicates that there is a communication problem with the server. The ' +
            'typical duration is within the range between 10 and 30 seconds.'
        )
    }

    /**
     * Returns help tip for the heartbeat status.
     */
    heartbeatStatusHelptip(): string {
        if (!this.serverStatus.commInterrupted || this.serverStatus.commInterrupted < 0) {
            return 'Status of the heartbeat communication with the ' + this.serverName + ' server is unknown.'
        } else if (this.serverStatus.commInterrupted > 0) {
            return (
                'Heartbeat communication with the ' +
                this.serverName +
                ' server ' +
                ' is interrupted. It means that the server has been failing to ' +
                ' respond to the ha-heartbeat commands longer than the configured ' +
                ' value of max-response-delay.'
            )
        }
        return 'The server responds to the ha-heartbeat commands sent by the ' + ' partner.'
    }

    /**
     * Returns help tip for the number of unacked clients.
     */
    unackedClientsHelptip(): string {
        return (
            'This is the number of clients considered unacked by the partner. ' +
            'This value is only set when the partner lost heartbeat communication ' +
            'with this server and started the failover procedure by monitoring ' +
            'whether the server is responding to the DHCP traffic. The unacked ' +
            'is the client which has been trying to get a lease from this server ' +
            'longer than the time specified with the max-ack-delay configuration ' +
            'parameter.'
        )
    }

    /**
     * Returns help tip for the number of connecting clients counted by
     * the partner server when the heartbeat communication between them is
     * interrupted.
     */
    connectingClientsHelptip(): string {
        return (
            'This is the total number of clients trying to get new lease ' +
            'from the server with which the partner server is unable to ' +
            'communicate via heartbeat. It includes both unacked clients ' +
            'and the clients which secs field or elapsed time option is ' +
            'below the max-ack-delay.'
        )
    }

    /**
     * Returns help tip for the number of packets directed to the server when
     * the heartbeat communication to this server gets interrupted.
     */
    analyzedPacketsHelptip(): string {
        return (
            'This is the total number of packets directed to the server ' +
            'with which the partner is unable to communicate via heartbeat. ' +
            'This may include several packets from the same client which ' +
            'retried to send DHCPDISCOVER or Solicit when the server failed to ' +
            'respond to the previous queries.'
        )
    }

    /**
     * Updates the panel colors according to the status fetched.
     *
     * The colors reflect the state of the HA. The green panel color
     * indicates that the server is in a desired state. The orange
     * color of the panel indicates that some abnormal situation has
     * occurred but it is not severe. For example, one of the servers
     * is down but the orange colored server has taken over serving the
     * DHCP clients. The red colored panel indicates an error which
     * most likely requires Administrator's action. For example, the
     * DHCP server has crashed.
     */
    private refreshPanelColors() {
        switch (this.serverWarnLevel()) {
            case 'ok':
                this.statusPanelClass = 'green-colored-panel'
                break
            case 'warn':
                this.statusPanelClass = 'orange-colored-panel'
                break
            default:
                this.statusPanelClass = 'red-colored-panel'
                break
        }
    }

    /**
     * Checks if the extended HA information is available for the given
     * Kea server.
     *
     * The extended information is returned since Kea 1.7.8 release. It
     * includes information about the failover progress, i.e. how many
     * clients have been trying to get the lease since the heartbeat
     * failure, how many clients failed to get the lease (unacked clients),
     * how many packets have been analyzed by the partner server etc.
     *
     * @returns true if the extended server status is supported, false
     *          otherwise.
     */
    extendedFormatSupported(): boolean {
        // Negative value of the commInterrupted is explicitly indicating
        // that the extended format is not supported. A zero or positive
        // value indicates it is supported.
        return this.hasStatus() && (!this.serverStatus.commInterrupted || this.serverStatus.commInterrupted >= 0)
    }

    /**
     * Checks what icon should be returned for the server.
     *
     * During normal operation the check icon is displayed. If the server is
     * unavailable the red exclamation mark is shown. In other cases a warning
     * exclamation mark on orange triangle is shown.
     */
    serverWarnLevel(): string {
        if (this.stateOk()) {
            return 'ok'
        }
        if (this.serverStatus.state === 'unavailable' || this.serverStatus.inTouch === false) {
            return 'error'
        }
        return 'warn'
    }

    /**
     * Checks if the state of the server is good.
     *
     * The desired state is either load-balancing, hot-standby or passive-backup.
     * In other cases it means that the servers are booting up or there is some
     * issue with the partner causing the server to go to partner-down.
     *
     * @returns true if the server is in the load-balancing, hot-standby or
     *          passive-backup state. This is used in the UI to highlight a
     *          potential problem.
     */
    stateOk(): boolean {
        return (
            this.hasStatus() &&
            (this.serverStatus.state === 'load-balancing' ||
                this.serverStatus.state === 'hot-standby' ||
                this.serverStatus.state === 'passive-backup')
        )
    }

    /**
     * Returns a comma separated list of HA scopes served by the server.
     *
     * This string is printed in the UI in the local server status box.
     *
     * @returns string containing comma separated list of scopes or (none).
     */
    formattedLocalScopes(): string {
        let scopes: string
        if (this.hasStatus() && this.serverStatus.scopes) {
            scopes = this.serverStatus.scopes.join(', ')
        }

        return scopes || '(none)'
    }

    /**
     * Returns formatted value of age.
     *
     * The age indicates how long ago the status of one of the servers has
     * been fetched. It is expressed in seconds. This function displays the
     * age in seconds for the age below 1 minute. It displays the age in
     * minutes otherwise. The nagative age value means that the age is not
     * yet determined in which case 'n/a' is displayed.
     *
     * @param age in seconds.
     * @returns string containing formatted age.
     */
    formattedAge(age): string {
        if (age && age < 0) {
            return 'n/a'
        }
        if (!age || age === 0) {
            return 'just now'
        }
        if (age < 60) {
            return age + ' seconds ago'
        }
        return Math.round(age / 60) + ' minutes ago'
    }

    /**
     * Returns formatted status of the control channel.
     *
     * Depending on the value of the boolean parameter specified, this function
     * returns the word "online" or "offline" to indicate the status of the
     * communication with one of the servers.
     *
     * @returns the descriptive information whether the server seems to be
     *          online or offline.
     */
    formattedControlStatus(): string {
        if (!this.hasStatus() || this.serverStatus.inTouch === null) {
            return 'unknown'
        }
        if (this.serverStatus.inTouch) {
            return 'online'
        }
        return 'offline'
    }

    /**
     * Returns timestamp as local time or 'n/a'.
     *
     * @param t Time value to be converted.
     *
     * @returns Time in local time or 'n/a' if the timestamp is equal to 0.
     */
    formattedTime(t): string {
        const localTime = datetimeToLocal(t)
        if (localTime.length === 0) {
            return 'n/a'
        }
        return localTime
    }

    /**
     * Returns formatted heartbeat status for the server.
     *
     * This information is only available if the extended status format
     * is supported (Kea 1.7.8 and later).
     *
     * @returns 'unknown' if extended format is not supported for this server,
     *          'ok' if the server is responding to the heartbeats,
     *          'failed' otherwise.
     */
    formattedHeartbeatStatus(): string {
        if (this.serverStatus.commInterrupted < 0) {
            return 'unknown'
        } else if (this.serverStatus.commInterrupted > 0) {
            return 'failed'
        }
        return 'ok'
    }

    /**
     * Returns formatted number of unacked clients for the server which fails
     * to respond to the heartbeats.
     *
     * The partner server starts to monitor the DHCP traffic directed to the
     * partner when the heartbeat has been failing with this server longer than
     * the configured period of time. The partner monitors the traffic by checking
     * the value of the 'secs' field (DHCPv4) or 'elapsed time' option (DHCPv6).
     * If these values exceed the configured threshold the client sending the
     * packet is considered unacked. If the number of unacked clients exceeds
     * the configured threshold for the number of unacked clients, the surviving
     * server enters the partner-down state.
     *
     * The string returned by this function includes the number of unacked
     * clients and the configured threshold for this number, e.g. 3 of 5,
     * which indicates that 3 out of 5 clients have been unacked so far.
     *
     * @return A string containing the number of unacked clients by the server
     *         and the maximum number of unacked clients before the server
     *         transitions to the partner-down state.
     */
    formattedUnackedClients(): string {
        let s = 'n/a'
        let unacked = 0
        let all = 0
        // Monitor unacked clients only if we're in the communication interrupted
        // state, i.e. heartbeat communication has been failing for a certain
        // period of time.
        if (this.hasStatus && this.serverStatus.commInterrupted > 0) {
            if (this.serverStatus.unackedClients != null) {
                unacked = this.serverStatus.unackedClients
            }
            all = unacked
            if (this.serverStatus.unackedClientsLeft != null) {
                all += this.serverStatus.unackedClientsLeft
            }
        }
        // If both unacked and unacked left values are 0 there is nothing
        // to print. It looks that we don't monitor unacked clients for
        // this server.
        if (all > 0) {
            s = unacked + ' of ' + (all + 1)
        }
        return s
    }

    /**
     * Returns formatted failover related number.
     *
     * This function is used to format the number of connecting clients or
     * analyzed packets. If the communication is not interrupted the returned
     * value is n/a. Otherwise, the number is returned.
     *
     * @returns A given value or n/a string.
     */
    formattedFailoverNumber(n): any {
        if (!this.serverStatus.unackedClients && !this.serverStatus.unackedClientsLeft) {
            return 'n/a'
        }
        if (n && n > 0) {
            return n
        }
        return 0
    }
}
