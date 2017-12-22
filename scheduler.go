package main

import (
	mesos "github.com/mesos/go-proto/mesos/v1"
)

// Interface for connecting a scheduler to Mesos. This
// interface is used both to manage the scheduler's lifecycle (start
// it, stop it, or wait for it to finish) and to interact with Mesos
// (e.g., launch tasks, kill tasks, etc.).
// See the MesosSchedulerDriver type for a concrete
// impl of a SchedulerDriver.
type SchedulerDriver interface {
	// Starts the scheduler driver. This needs to be called before any
	// other driver calls are made.
	Start() error

	// Stops the scheduler driver.
	Stop()

	// Launches the given set of tasks. Any resources remaining (i.e.,
	// not used by the tasks or their executors) will be considered
	// declined. The specified filters are applied on all unused
	// resources (see mesos.proto for a description of Filters).
	// Available resources are aggregated when mutiple offers are
	// provided. Note that all offers must belong to the same slave.
	// Invoking this function with an empty collection of tasks declines
	// offers in their entirety (see Scheduler::declineOffer).
	LaunchTasks(offerIDs []*mesos.OfferID, tasks []*mesos.TaskInfo, filters *mesos.Filters) error

	// Kills the specified task. Note that attempting to kill a task is
	// currently not reliable. If, for example, a scheduler fails over
	// while it was attempting to kill a task it will need to retry in
	// the future. Likewise, if unregistered / disconnected, the request
	// will be dropped (these semantics may be changed in the future).
	KillTask(taskID, agentID string) error

	// Declines an offer in its entirety and applies the specified
	// filters on the resources (see mesos.proto for a description of
	// Filters). Note that this can be done at any time, it is not
	// necessary to do this within the Scheduler::resourceOffers
	// callback.
	DeclineOffer(offerID *mesos.OfferID, filters *mesos.Filters) error

	// Removes all filters previously set by the framework (via
	// LaunchTasks()). This enables the framework to receive offers from
	// those filtered slaves.
	ReviveOffers() error

	// Allows the scheduler to query the status for non-terminal tasks.
	// This causes the master to send back the latest task status for
	// each task in 'tasks', if possible. Tasks that are no longer known
	// will result in a TASK_LOST, TASK_UNKNOWN, or TASK_UNREACHABLE update.
	// If 'tasks' is empty, then the master will send the latest status
	// for each task currently known.
	ReconcileTasks() error

	// Acknowledges the receipt of status update. Schedulers are
	// responsible for explicitly acknowledging the receipt of status
	// updates that have 'Update.status().uuid()' field set. Such status
	// updates are retried by the agent until they are acknowledged by
	// the scheduler.

	Acknowledge(status *mesos.TaskStatus) error
}

// Scheduler a type with callback attributes to be provided by frameworks
// schedulers.
//
// Each callback includes a reference to the scheduler driver that was
// used to run this scheduler. The pointer will not change for the
// duration of a scheduler (i.e., from the point you do
// SchedulerDriver.Start() to the point that SchedulerDriver.Stop()
// returns). This is intended for convenience so that a scheduler
// doesn't need to store a reference to the driver itself.
type Scheduler interface {

	// Invoked when the scheduler successfully registers with a Mesos
	// master. A unique ID (generated by the master) used for
	// distinguishing this framework from others and MasterInfo
	// with the ip and port of the current master are provided as arguments.
	Registered(SchedulerDriver, *mesos.FrameworkID, *mesos.MasterInfo)

	// Invoked when the scheduler receive "heartbeat" from the master.
	HeartBeated()

	// Invoked when resources have been offered to this framework. A
	// single offer will only contain resources from a single slave.
	// Resources associated with an offer will not be re-offered to
	// _this_ framework until either (a) this framework has rejected
	// those resources (see SchedulerDriver::launchTasks) or (b) those
	// resources have been rescinded (see Scheduler::offerRescinded).
	// Note that resources may be concurrently offered to more than one
	// framework at a time (depending on the allocator being used). In
	// that case, the first framework to launch tasks using those
	// resources will be able to use them while the other frameworks
	// will have those resources rescinded (or if a framework has
	// already launched tasks with those resources then those tasks will
	// fail with a TASK_LOST status and a message saying as much).
	ResourceOffers(SchedulerDriver, []*mesos.Offer)

	// Invoked when an offer is no longer valid (e.g., the slave was
	// lost or another framework used resources in the offer). If for
	// whatever reason an offer is never rescinded (e.g., dropped
	// message, failing over framework, etc.), a framwork that attempts
	// to launch tasks using an invalid offer will receive TASK_LOST
	// status updates for those tasks (see Scheduler::resourceOffers).
	OfferRescinded(SchedulerDriver, *mesos.OfferID)

	// Invoked when the status of a task has changed (e.g., a slave is
	// lost and so the task is lost, a task finishes and an executor
	// sends a status update saying so, etc). Note that returning from
	// this callback _acknowledges_ receipt of this status update! If
	// for whatever reason the scheduler aborts during this callback (or
	// the process exits) another status update will be delivered (note,
	// however, that this is currently not true if the slave sending the
	// status update is lost/fails during that time).
	StatusUpdate(SchedulerDriver, *mesos.TaskStatus)
}
