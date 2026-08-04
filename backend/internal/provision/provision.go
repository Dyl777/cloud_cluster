package provision

// Instance represents a rented GPU unit.
type Instance struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	GPUName  string `json:"gpu_name"`
	NumGPUs  int    `json:"num_gpus"`
	Status   string `json:"status"`
	Provider string `json:"provider"`
}
