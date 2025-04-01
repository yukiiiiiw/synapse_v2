package enum

// DeploymentStatus the deployment lifecycle status.
type DeploymentStatus int

const (
	Deployment_Status_Failure   DeploymentStatus = -1 // Failure
	Deployment_Status_Deploying DeploymentStatus = 0  // Deploying
	Deployment_Status_Updating  DeploymentStatus = 10 // Updating
	Deployment_Status_Deleting  DeploymentStatus = 11 // Deleting
	Deployment_Status_Pausing   DeploymentStatus = 12 // Pausing
	Deployment_Status_Running   DeploymentStatus = 20 // Running
	Deployment_Status_Deleted   DeploymentStatus = 21 // Deleted
	Deployment_Status_Paused    DeploymentStatus = 22 // Paused
)
