package orchestrator

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func (s *OrchestratorService) ExecuteOnVirtualMachine(ctx basecontext.ApiContext, vmId string, request models.VirtualMachineExecuteCommandRequest) (*models.VirtualMachineExecuteCommandResponse, error) {
	vm, err := s.GetVirtualMachine(ctx, vmId)
	if err != nil {
		return nil, err
	}
	if vm == nil {
		return nil, errors.NewWithCodef(404, "Virtual machine %s not found", vmId)
	}

	host, err := s.GetHost(ctx, vm.HostId)
	if err != nil {
		return nil, err
	}
	if host == nil {
		return nil, errors.NewWithCodef(404, "Host %s not found", vm.HostId)
	}

	if !host.Enabled {
		return nil, errors.NewWithCodef(400, "Host %s is disabled", host.Host)
	}
	if host.State != "healthy" {
		return nil, errors.NewWithCodef(400, "Host %s is not healthy", host.Host)
	}

	return s.ExecuteOnHostVirtualMachine(ctx, vm.HostId, vm.ID, request)
}

func (s *OrchestratorService) ExecuteOnHostVirtualMachine(ctx basecontext.ApiContext, hostId string, vmId string, request models.VirtualMachineExecuteCommandRequest) (*models.VirtualMachineExecuteCommandResponse, error) {
	vm, err := s.GetVirtualMachine(ctx, vmId)
	if err != nil {
		return nil, err
	}
	if vm == nil {
		return nil, errors.NewWithCodef(404, "Virtual machine %s not found", vmId)
	}

	host, err := s.GetHost(ctx, hostId)
	if err != nil {
		return nil, err
	}
	if host == nil {
		return nil, errors.NewWithCodef(404, "Host %s not found", hostId)
	}

	if !host.Enabled {
		return nil, errors.NewWithCodef(400, "Host %s is disabled", host.Host)
	}
	if host.State != "healthy" {
		return nil, errors.NewWithCodef(400, "Host %s is not healthy", host.Host)
	}

	currentVmState, err := s.GetHostVirtualMachineStatus(ctx, host.ID, vm.ID)
	if err != nil {
		return nil, err
	}
	vm.State = currentVmState.Status

	if vm.State != "running" {
		return nil, errors.NewWithCodef(400, "Virtual machine %s is not running", vmId)
	}

	return s.CallExecuteOnHostVirtualMachine(host, vm.ID, request)
}

func (s *OrchestratorService) CallExecuteOnHostVirtualMachine(host *data_models.OrchestratorHost, vmId string, request models.VirtualMachineExecuteCommandRequest) (*models.VirtualMachineExecuteCommandResponse, error) {
	httpClient := s.getApiClient(*host)
	path := "/machines/" + vmId + "/execute"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	var response models.VirtualMachineExecuteCommandResponse
	_, err = httpClient.Put(url.String(), request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
