package typhoon

import "github.com/TyphoonMC/TyphoonCore/blocks"

var(
	BLOCK_REGISTRY = generateRegistry()
)

type BlockRegistry struct {
	counter uint16
	NameToGuid map[string]uint16
	GuidToName map[uint16]string
}

func generateRegistry() *BlockRegistry {
	r := &BlockRegistry{
		counter:    0,
		NameToGuid: make(map[string]uint16),
		GuidToName: make(map[uint16]string),
	}

	for _, v := range blocks.GetLegacyMapping() {
		r.GetGuid(v)
	}

	return r
}

func (registry *BlockRegistry) GetGuid(name string) uint16 {
	if val, ok := registry.NameToGuid[name]; ok {
		return val
	} else {
		nid := registry.counter
		registry.counter++
		registry.NameToGuid[name] = nid
		registry.GuidToName[nid] = name
		return nid
	}
}

func (registry *BlockRegistry) GetName(guid uint16) string {
	if val, ok := registry.GuidToName[guid]; ok {
		return val
	} else {
		return "minecraft:stone"
	}
}

func (registry *BlockRegistry) GetBlockId(name string, proto Protocol) int {
	if proto >= V1_13 {
		return blocks.GetV1_13FromName(name)
	}
	return blocks.GetLegacyFromName(name)
}