package compat

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gpuhub/cloud/internal/marketplace"
	"github.com/gpuhub/cloud/internal/provision"
	"github.com/gpuhub/cloud/internal/shared/httpx"
)

// createInstance mirrors PUT /asks/{id}/ — launch an instance from an offer.
func (s *Server) createInstance(w http.ResponseWriter, r *http.Request) {
	offerID := r.PathValue("id")
	o, ok := s.offerByID(offerID)
	if !ok {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"success": false, "msg": "no such ask: " + offerID})
		return
	}
	var body struct {
		Image    string            `json:"image"`
		Env      map[string]string `json:"env"`
		Disk     float64           `json:"disk"`
		Label    string            `json:"label"`
		Onstart  string            `json:"onstart"`
		Runtype  string            `json:"runtype"`
		Price    float64           `json:"price"`
		ClientID string            `json:"client_id"`
	}
	_ = httpx.Decode(r, &body)
	if body.Image == "" {
		body.Image = "pytorch/pytorch:latest"
	}
	if body.Price <= 0 {
		body.Price = o.Dph
	}
	inst := provision.Instance{
		ID:        offerID,
		UserID:    s.userID,
		GPUName:   o.GPUName,
		GPUVRAM:   o.GPUVRAM,
		NumGPUs:   o.NumGPUs,
		CPU:       o.CPU,
		RAM:       o.RAM,
		Disk:      body.Disk,
		Region:    o.Region,
		Provider:  o.Provider,
		Image:     body.Image,
		Label:     body.Label,
		Status:    "running",
		Price:     body.Price,
		SSHPort:   10000 + len(s.prov.List("")) + 1,
		PublicIP:  "35.10.0." + strconv.Itoa(200+len(s.prov.List(""))%50),
		StartDate: time.Now().Unix(),
	}
	s.prov.Launch(inst)
	// mark offer rented so it leaves the search pool
	s.markRented(offerID, true)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"new_contract": map[string]any{
			"id": offerID, "machine_id": offerID,
		},
		"instance": instanceRow(inst),
	})
}

// createInstancesBulk mirrors POST /asks/bulk/.
func (s *Server) createInstancesBulk(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs   []int  `json:"ids"`
		Image string `json:"image"`
		Disk  int    `json:"disk"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "msg": err.Error()})
		return
	}
	if len(body.IDs) == 0 {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "msg": "no ids"})
		return
	}
	created := make([]map[string]any, 0, len(body.IDs))
	for _, oid := range body.IDs {
		offerID := strconv.Itoa(oid)
		o, ok := s.offerByID(offerID)
		if !ok {
			continue
		}
		inst := provision.Instance{
			ID: offerID, UserID: s.userID, GPUName: o.GPUName, GPUVRAM: o.GPUVRAM,
			NumGPUs: o.NumGPUs, CPU: o.CPU, RAM: o.RAM, Disk: float64(body.Disk),
			Region: o.Region, Provider: o.Provider, Image: body.Image,
			Status: "running", Price: o.Dph, StartDate: time.Now().Unix(),
		}
		s.prov.Launch(inst)
		created = append(created, map[string]any{"id": offerID, "machine_id": offerID})
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "new_contracts": created})
}

// showInstances mirrors GET /instances (v0, returns {"instances": [...]}).
func (s *Server) showInstances(w http.ResponseWriter, r *http.Request) {
	rows := make([]map[string]any, 0)
	for _, inst := range s.prov.List(s.userID) {
		rows = append(rows, instanceRow(inst))
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"instances": rows})
}

// showInstancesV1 mirrors GET /api/v1/instances/ (paginated instance rows).
func (s *Server) showInstancesV1(w http.ResponseWriter, r *http.Request) {
	rows := make([]map[string]any, 0)
	for _, inst := range s.prov.List(s.userID) {
		rows = append(rows, instanceRow(inst))
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"instances": rows, "next_token": nil,
		"total_instances": len(rows), "label_counts": map[string]any{},
	})
}

// showInstance mirrors GET /instances/{id}/?owner=me.
func (s *Server) showInstance(w http.ResponseWriter, r *http.Request) {
	inst, ok := s.prov.Get(r.PathValue("id"))
	if !ok {
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"instances": nil})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"instances": instanceRow(inst)})
}

// updateInstance mirrors PUT /instances/{id}/ (label / origin).
func (s *Server) updateInstance(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Label string `json:"label"`
		Image string `json:"image"`
	}
	_ = httpx.Decode(r, &body)
	id := r.PathValue("id")
	if body.Label != "" {
		if _, ok := s.prov.Update(id, func(i *provision.Instance) { i.Label = body.Label }); !ok {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"success": false, "msg": "no such instance"})
			return
		}
	}
	if _, ok := s.prov.Update(id, func(i *provision.Instance) { i.Image = body.Image }); !ok && body.Label == "" {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"success": false, "msg": "no such instance"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"success": true})
}

// updateInstances mirrors PUT /instances/ (bulk start/stop by state).
func (s *Server) updateInstances(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs   []int  `json:"ids"`
		State string `json:"state"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "msg": err.Error()})
		return
	}
	for _, oid := range body.IDs {
		idStr := strconv.Itoa(oid)
		s.prov.Update(idStr, func(i *provision.Instance) { i.Status = body.State })
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"success": true})
}

// destroyInstances mirrors DELETE /instances/ (bulk).
func (s *Server) destroyInstances(w http.ResponseWriter, r *http.Request) {
	var body struct {
		InstanceIDs []int `json:"instance_ids"`
		IDs         []int `json:"ids"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "msg": err.Error()})
		return
	}
	ids := body.InstanceIDs
	if len(ids) == 0 {
		ids = body.IDs
	}
	for _, oid := range ids {
		idStr := strconv.Itoa(oid)
		if s.prov.Destroy(idStr) {
			s.markRented(idStr, false)
		}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"success": true})
}

// destroyInstance mirrors DELETE /instances/{id}/.
func (s *Server) destroyInstance(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if !s.prov.Destroy(idStr) {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"success": false, "msg": "no such instance"})
		return
	}
	s.markRented(idStr, false)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"success": true})
}

// rebootInstance mirrors PUT /instances/reboot/{id}/.
func (s *Server) rebootInstance(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if _, ok := s.prov.Update(idStr, func(i *provision.Instance) { i.Status = "running" }); !ok {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"success": false, "msg": "no such instance"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"success": true})
}

// commandInstance mirrors PUT /instances/command/{id}/ (exec on instance).
func (s *Server) commandInstance(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Command string `json:"command"`
	}
	_ = httpx.Decode(r, &body)
	_ = r.PathValue("id")
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true, "result": "(mock) executed: " + body.Command,
	})
}

// instanceBalance mirrors GET /instances/balance/{id}/.
func (s *Server) instanceBalance(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	inst, ok := s.prov.Get(idStr)
	if !ok {
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"success": false, "msg": "no such instance"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"balance": map[string]any{
			"credits_used": 0.0, "dph_total": inst.Price,
			"id": idStr, "status": inst.Status,
		},
	})
}

func instanceRow(i provision.Instance) map[string]any {
	extraEnv := make([][2]string, 0)
	duration := time.Now().Unix() - i.StartDate
	return map[string]any{
		"id": i.ID, "machine_id": i.ID, "host_id": i.Provider,
		"gpu_name": i.GPUName, "gpu_ram": i.GPUVRAM, "num_gpus": i.NumGPUs,
		"cpu_cores": i.CPU, "cpu_ram": i.RAM, "disk_space": i.Disk,
		"dph_total": i.Price, "display_price": i.Price,
		"actual_status": i.Status, "status_msg": "Instance running",
		"start_date": i.StartDate, "duration": duration,
		"ssh_port": i.SSHPort, "public_ipaddr": i.PublicIP,
		"image_uuid": i.Image, "label": i.Label,
		"extra_env": extraEnv, "region": i.Region, "provider": i.Provider,
		"owner": i.UserID,
	}
}

// offerByID finds a market offer by id.
func (s *Server) offerByID(idStr string) (marketplace.Offer, bool) {
	for _, o := range s.reg.ListAll() {
		if o.ID == idStr {
			return o, true
		}
	}
	return marketplace.Offer{}, false
}

// markRented flips the rented bit on the underlying market offer so instances
// in flight disappear from further searches (mirrors calibration behavior).
func (s *Server) markRented(idStr string, rented bool) {
	// The registry is currently a read-only mock snapshot; rented state is
	// tracked via provision instances instead. Expansive adapter V2 will push
	// this through the provider. Keep the hook for parity clarity.
	_ = idStr
	_ = rented
}
