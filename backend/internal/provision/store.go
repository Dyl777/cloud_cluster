package provision

import (
	"database/sql"
	"errors"
)

// ErrNotFound is returned by a Store when an instance is absent.
var ErrNotFound = errors.New("provision: instance not found")

// InstanceStore persists rented GPU instances. memoryStore keeps the default
// single-process behavior; pgStore persists to Postgres.
type InstanceStore interface {
	Put(i *Instance) error
	Get(id string) (*Instance, error)
	Delete(id string) error
	List(userID string) ([]Instance, error)
}

// memoryStore implements InstanceStore in-process.
type memoryStore struct{ instances map[string]Instance }

func newMemoryStore() *memoryStore {
	return &memoryStore{instances: make(map[string]Instance)}
}

func (m *memoryStore) Put(i *Instance) error {
	m.instances[i.ID] = *i
	return nil
}

func (m *memoryStore) Get(id string) (*Instance, error) {
	i, ok := m.instances[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &i, nil
}

func (m *memoryStore) Delete(id string) error {
	if _, ok := m.instances[id]; !ok {
		return ErrNotFound
	}
	delete(m.instances, id)
	return nil
}

func (m *memoryStore) List(userID string) ([]Instance, error) {
	out := []Instance{}
	for _, i := range m.instances {
		if userID != "" && i.UserID != userID {
			continue
		}
		out = append(out, i)
	}
	return out, nil
}

// pgStore implements InstanceStore on Postgres.
type pgStore struct{ db *sql.DB }

// NewPGStore returns provision persistence backed by Postgres.
func NewPGStore(db *sql.DB) InstanceStore { return &pgStore{db: db} }

const instanceCols = "id, user_id, gpu_name, gpu_vram, num_gpus, cpu_cores, cpu_ram, " +
	"disk_space, region, provider, image, label, status, price, ssh_port, public_ip, start_date"

func scanInstance(row interface{ Scan(...any) error }) (*Instance, error) {
	var i Instance
	err := row.Scan(&i.ID, &i.UserID, &i.GPUName, &i.GPUVRAM, &i.NumGPUs, &i.CPU, &i.RAM,
		&i.Disk, &i.Region, &i.Provider, &i.Image, &i.Label, &i.Status, &i.Price,
		&i.SSHPort, &i.PublicIP, &i.StartDate)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (p *pgStore) Put(i *Instance) error {
	_, err := p.db.Exec(`INSERT INTO provision_instance (`+instanceCols+`, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17, now(), now())
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			gpu_name = EXCLUDED.gpu_name,
			gpu_vram = EXCLUDED.gpu_vram,
			num_gpus = EXCLUDED.num_gpus,
			cpu_cores = EXCLUDED.cpu_cores,
			cpu_ram = EXCLUDED.cpu_ram,
			disk_space = EXCLUDED.disk_space,
			region = EXCLUDED.region,
			provider = EXCLUDED.provider,
			image = EXCLUDED.image,
			label = EXCLUDED.label,
			status = EXCLUDED.status,
			price = EXCLUDED.price,
			ssh_port = EXCLUDED.ssh_port,
			public_ip = EXCLUDED.public_ip,
			start_date = EXCLUDED.start_date,
			updated_at = now()`,
		i.ID, i.UserID, i.GPUName, i.GPUVRAM, i.NumGPUs, i.CPU, i.RAM, i.Disk, i.Region,
		i.Provider, i.Image, i.Label, i.Status, i.Price, i.SSHPort, i.PublicIP, i.StartDate)
	return err
}

func (p *pgStore) Get(id string) (*Instance, error) {
	return scanInstance(p.db.QueryRow(
		"SELECT "+instanceCols+" FROM provision_instance WHERE id = $1", id))
}

func (p *pgStore) Delete(id string) error {
	res, err := p.db.Exec("DELETE FROM provision_instance WHERE id = $1", id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *pgStore) List(userID string) ([]Instance, error) {
	query := "SELECT " + instanceCols + " FROM provision_instance"
	args := []any{}
	if userID != "" {
		query += " WHERE user_id = $1"
		args = append(args, userID)
	}
	query += " ORDER BY created_at, id"
	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Instance{}
	for rows.Next() {
		i, err := scanInstance(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *i)
	}
	return out, rows.Err()
}

var _ InstanceStore = (*memoryStore)(nil)
var _ InstanceStore = (*pgStore)(nil)