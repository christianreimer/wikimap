package tiles

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func WriteTile(dir string, tile GeneratedTile) error {
	tileDir := filepath.Join(dir, fmt.Sprintf("%d", tile.Z), fmt.Sprintf("%d", tile.X))
	if err := os.MkdirAll(tileDir, 0o755); err != nil {
		return err
	}

	path := filepath.Join(tileDir, fmt.Sprintf("%d.pbf", tile.Y))
	data, err := json.Marshal(tile)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func ReadTile(path string) (GeneratedTile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return GeneratedTile{}, err
	}

	var tile GeneratedTile
	err = json.Unmarshal(data, &tile)
	return tile, err
}
