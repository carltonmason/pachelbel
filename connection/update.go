package connection

import (
	"fmt"

	compose "github.com/benjdewan/gocomposeapi"
)

func update(cxn *Connection, accessor Accessor) error {
	deployment := accessor.(Deployment)
	existing, ok := cxn.getDeploymentByName(deployment.GetName())
	if !ok {
		return fmt.Errorf("Attempting to update '%s', but it doesn't exist",
			deployment.GetName())
	}

	if err := updateScalings(cxn, existing.ID, deployment); err != nil {
		return err
	}

	if err := updateNotes(cxn, existing.ID, deployment); err != nil {
		return err
	}

	if err := updateVersion(cxn, existing, deployment); err != nil {
		return err
	}

	if err := addTeamRoles(cxn, existing.ID, deployment.GetTeamRoles()); err != nil {
		return err
	}

	// Changing versions and sizes can change the deployment ID. Ensure
	// we have the latest/live value
	updatedDeployment, errs := cxn.client.GetDeploymentByName(existing.Name)
	if len(errs) != 0 {
		return fmt.Errorf("Unable to get deployment information for '%s':\n%v",
			deployment.GetName(), errs)
	}
	cxn.newDeploymentIDs.Store(updatedDeployment.ID, struct{}{})
	return nil
}

func updateScalings(cxn *Connection, id string, deployment Deployment) error {
	existingScalings, errs := cxn.client.GetScalings(id)
	if len(errs) != 0 {
		return fmt.Errorf("Unable to get current scaling for '%s':\n%v",
			deployment.GetName(), errs)
	}

	if existingScalings.AllocatedUnits == deployment.GetScaling() {
		return nil
	}

	recipe, errs := cxn.client.SetScalings(compose.ScalingsParams{
		DeploymentID: id,
		Units:        deployment.GetScaling(),
	})
	if len(errs) != 0 {
		return fmt.Errorf("Unable to resize '%s':\n%v",
			deployment.GetName(), errs)
	}

	return cxn.waitOnRecipe(recipe.ID, deployment.GetTimeout())
}

func updateNotes(cxn *Connection, id string, deployment Deployment) error {
	if len(deployment.GetNotes()) == 0 {
		return nil
	}

	_, errs := cxn.client.PatchDeployment(compose.PatchDeploymentParams{
		DeploymentID: id,
		Notes:        deployment.GetNotes(),
	})
	if len(errs) != 0 {
		return fmt.Errorf("Unable to update notes for '%s':\n%v",
			id, errs)
	}
	return nil
}

func updateVersion(cxn *Connection, existing *compose.Deployment, deployment Deployment) error {
	if len(deployment.GetVersion()) == 0 || existing.Version == deployment.GetVersion() {
		return nil
	}

	transitions, errs := cxn.client.GetVersionsForDeployment(existing.ID)
	if len(errs) != 0 || transitions == nil {
		return fmt.Errorf("Error fetching upgrade information for '%s':\n%v",
			existing.Name, errs)
	}

	err := validTransition(*transitions, deployment)
	if err != nil {
		return err
	}

	recipe, errs := cxn.client.UpdateVersion(existing.ID,
		deployment.GetVersion())
	if errs != nil {
		return fmt.Errorf("Unable to upgrade '%s' to version '%s':\n%v",
			deployment.GetName(), deployment.GetVersion(), errs)
	}

	return cxn.waitOnRecipe(recipe.ID, deployment.GetTimeout())
}

func validTransition(transitions []compose.VersionTransition, deployment Deployment) error {
	for _, transition := range transitions {
		if transition.ToVersion == deployment.GetVersion() {
			return nil
		}
	}
	return fmt.Errorf("Cannot upgrade '%s' to version '%s'.",
		deployment.GetName(), deployment.GetVersion())
}

func dryRunUpdate(cxn *Connection, accessor Accessor) error {
	deployment, ok := cxn.getDeploymentByName(accessor.GetName())
	if !ok {
		// This should never happen
		panic("map integrity failure")
	}
	cxn.newDeploymentIDs.Store(deployment.ID, struct{}{})
	return nil
}
