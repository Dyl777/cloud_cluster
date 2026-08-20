package provision

// Instance represents a rented GPU unit.
type Instance struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	GPUName   string  `json:"gpu_name"`
	GPUVRAM   int64   `json:"gpu_ram"`
	NumGPUs   int     `json:"num_gpus"`
	CPU       int     `json:"cpu_cores"`
	RAM       int     `json:"cpu_ram"`
	Disk      float64 `json:"disk_space"`
	Region    string  `json:"region"`
	Provider  string  `json:"provider"`
	Image     string  `json:"image"`
	Label     string  `json:"label"`
	Status    string  `json:"status"`
	Price     float64 `json:"dph_total"`
	SSHPort   int     `json:"ssh_port"`
	PublicIP  string  `json:"public_ipaddr"`
	StartDate int64   `json:"start_date"`
}
