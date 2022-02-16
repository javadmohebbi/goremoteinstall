package goremoteinstall

type Status uint

const (
	// Global Status
	STATUS_NOT_STARTED_YET Status = iota + 1000
	STATUS_STARTING
	STATUS_STARTED
	STATUS_WAITING
	STATUS_PENDING
	STATUS_FINISHED
	STATUS_FINISHED_OK
	STATUS_FINISHED_ERR

	// DIAL 445 status
	STATUS_DIAL_445
	STATUS_DIAL_445_FAILED
	STATUS_DIAL_445_OPEN
	STATUS_DIAL_445_OK

	// SMB SESSION status
	STATUS_PREPARE_SMB_SESSION
	STATUS_PREPARE_SMB_SESSION_LIST_SHARES
	STATUS_PREPARE_SMB_SESSION_FAILED
	STATUS_PREPARE_SMB_SESSION_DONE

	// SMB List status
	STATUS_PREPARE_SMB_SESSION_LIST_SHARES_FAILED
	STATUS_PREPARE_SMB_SESSION_LIST_SHARES_NOADMIN
	STATUS_PREPARE_SMB_SESSION_LIST_SHARES_ADMINOK
	STATUS_PREPARE_SMB_SESSION_LIST_SHARES_OK

	// SMB Copy status
	STATUS_SMB_COPY
	STATUS_SMB_MOUNT_ADMIN_FAILED
	STATUS_SMB_MKDIR_IN_ADMIN_FAILED
	STATUS_SMB_COPY_AGENT
	STATUS_SMB_COPY_AGENT_FAILED
	STATUS_SMB_COPY_AGENT_DONE
	STATUS_SMB_COPY_READ_PKG_DIR
	STATUS_SMB_COPY_READ_PKG_DIR_FAILED
	STATUS_SMB_FAILED
	STATUS_SMB_COPY_READ_PACKAGE_FILES
	STATUS_SMB_COPY_PKG_FILE_TO_TARGET
	STATUS_SMB_COPY_PKG_FILE_TO_TARGET_DONE
	STATUS_SMB_COPY_PKG_FILE_TO_TARGET_FAILED
	STATUS_SMB_DONE

	// Execute status
	STATUS_EXECUTING
	STATUS_EXECUTED
	STATUS_EXECUTING_FAILED
	STATUS_EXECITING_DONE
	STATUS_EXECITING_TIMEOUT

	// Windows installer status
	STATUS_CREATING_SERVICE
	STATUS_SERVICE_CREATION_FAILED
	STATUS_SERVICE_CREATED
	STATUS_STARTING_SERVICE
	STATUS_STARTING_SERVICE_FAILED
	STATUS_SERVICE_STARTED
	STATUS_STOPPING_SERVICE
	STATUS_STOPPING_SERVICE_FAILED
)

func (s Status) String() string {
	return [...]string{
		// Global Status
		"N/A",
		"starting",
		"started",
		"waiting",
		"pending",
		"finished",
		"successfully finished",
		"completed with error",

		// DIAL 445 status
		"checking tcp/445",
		"can not dial tcp/445",
		"tcp/445 is open",
		"tcp/445 is OK",

		// SMB SESSION status
		"preparing SMB session",
		"getting shared items",
		"preparing SMB session failed",
		"preparing SMB session is done",

		// SMB List status
		"listing shared items faild",
		"could not find ADMIN$ in the shared items",
		"ADMIN$ is listed on the shared items",
		"listing shared items is successfully completed",

		// SMB Copy status
		"start copying files over SMB to the target machine",
		"could not access to ADMIN$ shared folder",
		"could not create directory in ADMIN$",
		"copy installer agent",
		"copy installer agent executable failed",
		"copy installer agent executable done",
		"read package dir for the files to copy on target machines",
		"read package dir failed",
		"could not copy file(s) over SMB to target machine",
		"unable to read some of package files",
		"copy file to target",
		"copy file to target done",
		"copy file to target failed",
		"files are copied to the target machines",

		// Execute status
		"executing the bootstrap",
		"bootstrap executed",
		"executing the bootstrap failed",
		"the bootstrap executed successfully",
		"executing the bootstrap task has timed out",

		// Windows installer status
		"creating installer service",
		"creating installer service failed",
		"installer service created",
		"starting installer service",
		"starting installer service failed",
		"installer service started",
		"stopping installer service",
		"stopping installer service failed",
	}[s-1000]
}
