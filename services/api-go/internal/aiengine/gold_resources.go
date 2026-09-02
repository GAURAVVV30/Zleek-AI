package aiengine

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed data/gold_tier_resources.json
var goldTierJSON []byte

type GoldResource struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Provider    string `json:"provider"`
	Type        string `json:"type"` // "video", "documentation", "hands_on"
}

type GoldResourcesGroup struct {
	Documentation []GoldResource `json:"documentation"`
	Video         []GoldResource `json:"video"`
	HandsOn       []GoldResource `json:"hands_on"`
}

type GoldModule struct {
	ModuleID     string             `json:"module_id"`
	ModuleNumber int                `json:"module_number"`
	ModuleName   string             `json:"module_name"`
	Resources    GoldResourcesGroup `json:"resources"`
}

type GoldRole struct {
	RoleID       string       `json:"role_id"`
	RoleName     string       `json:"role_name"`
	PDFFilename  string       `json:"pdf_filename"`
	TotalModules int          `json:"total_modules"`
	Modules      []GoldModule `json:"modules"`
}

type GoldResourceLookup struct {
	Roles map[string]GoldRole
}

var (
	goldLookup *GoldResourceLookup
	goldOnce   sync.Once
)

func GetGoldResourceLookup() *GoldResourceLookup {
	goldOnce.Do(func() {
		var roles map[string]GoldRole
		if err := json.Unmarshal(goldTierJSON, &roles); err != nil {
			roles = make(map[string]GoldRole)
		}
		goldLookup = &GoldResourceLookup{Roles: roles}
	})
	return goldLookup
}

var domainUUIDMap = map[string]string{
	"f9ec7df2-79d6-52b1-9786-2be23e1738ee": "ai_data_scientist",
	"ffc78d4c-4ed7-5f25-a82c-52c38181bafd": "ai_engineer",
	"3e5fea96-c7e9-5817-a396-daacbbafeb7b": "backend_engineer",
	"c46c25c9-44c3-581e-b903-d217a9c8c03c": "data_analyst",
	"4a9bcbcf-0459-531f-856d-b20c270c376b": "full_stack",
	"66a1c5de-4d1a-5303-bce1-5439aad10da2": "devops_sre",
	"9190cc21-9768-5b94-850e-3a28f66c4055": "frontend_engineer",
	"5e728fa9-b31e-5170-9dce-dec79630dd21": "full_stack",
	"3f02aad3-d57b-5129-9cf2-b914ed7e313e": "machine_learning",
	"2e0c2a7f-e1e1-534c-8501-4074196f0915": "mobile_engineer",
	"714b074c-43b9-5ecc-861d-f1f3bab7663f": "product_manager",
	"1296678f-0912-5d9e-8d1a-5d331d2b1cd7": "software_architect",
}

func normalizeRoleID(roleID string) string {
	r := strings.TrimSpace(strings.ToLower(roleID))
	if mapped, ok := domainUUIDMap[r]; ok {
		return mapped
	}
	return r
}

func (g *GoldResourceLookup) GetRole(roleID string) (GoldRole, bool) {
	roleID = normalizeRoleID(roleID)
	if roleID == "data_engineer" {
		return GoldRole{}, false
	}
	role, ok := g.Roles[roleID]
	return role, ok
}

// GetGoldModuleResources returns gold tier resources for roleID + moduleQuery (ID, number, or title)
func (g *GoldResourceLookup) GetGoldModuleResources(roleID, moduleQuery string) (*GoldModule, bool) {
	roleID = normalizeRoleID(roleID)
	if roleID == "data_engineer" {
		return nil, false
	}
	role, ok := g.Roles[roleID]
	if !ok {
		return nil, false
	}

	moduleQueryLower := strings.ToLower(strings.TrimSpace(moduleQuery))

	// 1. Exact ID or Name match
	for _, m := range role.Modules {
		if strings.EqualFold(m.ModuleID, moduleQueryLower) || strings.EqualFold(m.ModuleName, moduleQueryLower) {
			return &m, true
		}
	}

	// 2. Numeric match (module_number OR 1-based index in PDF list)
	var modNum int
	if _, err := fmt.Sscanf(moduleQueryLower, "%d", &modNum); err == nil && modNum > 0 {
		// Check explicit ModuleNumber match
		for _, m := range role.Modules {
			if m.ModuleNumber == modNum {
				return &m, true
			}
		}
		// Check 1-based index match (e.g. 1st module in role's PDF list)
		if modNum <= len(role.Modules) {
			return &role.Modules[modNum-1], true
		}
	}

	// 3. Substring / Keyword title & ID match
	for _, m := range role.Modules {
		mNameLower := strings.ToLower(m.ModuleName)
		mIDLower := strings.ToLower(m.ModuleID)
		if strings.Contains(mIDLower, moduleQueryLower) || strings.Contains(mNameLower, moduleQueryLower) || strings.Contains(moduleQueryLower, mNameLower) {
			return &m, true
		}
	}

	// 4. Extract sequence number from node_id e.g. "da_01_analytics_foundations" -> 1
	var numStr string
	for _, part := range strings.Split(moduleQueryLower, "_") {
		if len(part) == 2 && part[0] >= '0' && part[0] <= '9' && part[1] >= '0' && part[1] <= '9' {
			numStr = part
			break
		}
	}
	if numStr != "" {
		var seqNum int
		fmt.Sscanf(numStr, "%d", &seqNum)
		if seqNum > 0 {
			for _, m := range role.Modules {
				if m.ModuleNumber == seqNum {
					return &m, true
				}
			}
			if seqNum <= len(role.Modules) {
				return &role.Modules[seqNum-1], true
			}
		}
	}

	// 5. Fallback to 1st module if query is '1' or '01' and role has modules
	if (moduleQueryLower == "1" || moduleQueryLower == "01") && len(role.Modules) > 0 {
		return &role.Modules[0], true
	}

	return nil, false
}
