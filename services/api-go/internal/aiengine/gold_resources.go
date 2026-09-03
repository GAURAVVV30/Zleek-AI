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

var nodeAliasMap = map[string]string{
	"ai_ds_01_python_foundations":                      "ai_data_scientist_m01_python_for_data_science_and_scientific_computing",
	"ai_ds_02_statistics_probability":                 "ai_data_scientist_m02_exploratory_data_analysis_eda_and_visualization",
	"ai_ds_03_data_engineering_basics":                "ai_data_scientist_m03_feature_engineering_and_data_preparation",
	"ai_ds_04_exploratory_data_analysis":               "ai_data_scientist_m02_exploratory_data_analysis_eda_and_visualization",
	"ai_ds_05_machine_learning_basics":                 "ai_data_scientist_m05_supervised_learning_algorithms_and_modeling",
	"ai_ds_06_feature_engineering":                     "ai_data_scientist_m03_feature_engineering_and_data_preparation",
	"ai_ds_07_model_evaluation":                        "ai_data_scientist_m04_model_evaluation_and_experimentation",
	"ai_ds_08_supervised_learning":                     "ai_data_scientist_m05_supervised_learning_algorithms_and_modeling",
	"ai_ds_09_unsupervised_learning":                   "ai_data_scientist_m06_unsupervised_learning_and_representation_learning",
	"ai_ds_10_deep_learning":                          "ai_data_scientist_m07_deep_learning_for_structured_and_unstructured_data",
	"ai_ds_11_mlops_and_deployment":                    "ai_data_scientist_m08_mlops_deployment_and_production_engineering",
	"ai_ds_12_research_design_and_communication":       "ai_data_scientist_m09_experiment_design_decision_making_and_ai_systems",
	"ai_eng_01_python_and_data_literacy":              "ai_engineer_m01_python_programming_and_data_literacy",
	"ai_eng_02_math_for_ml_and_ai":                     "ai_engineer_m02_mathematics_for_machine_learning_and_ai",
	"ai_eng_03_ml_foundations":                         "ai_engineer_m03_machine_learning_foundations",
	"ai_eng_04_feature_and_data_pipeline":              "ai_engineer_m04_feature_engineering_and_data_pipelines",
	"ai_eng_05_supervised_learning_algorithms":         "ai_engineer_m05_supervised_learning_and_ensemble_models",
	"ai_eng_06_model_evaluation_and_experimentation":  "ai_engineer_m06_model_evaluation_and_experimentation",
	"ai_eng_07_deep_learning_basics":                   "ai_engineer_m07_deep_learning_fundamentals",
	"ai_eng_08_nlp_and_llms":                           "ai_engineer_m08_natural_language_processing_and_large_language_models",
	"ai_eng_09_model_serving_and_inference":            "ai_engineer_m09_model_deployment_serving_and_orchestration",
	"ai_eng_10_mlops_and_observability":                "ai_engineer_m10_mlops_monitoring_and_continuous_integration",
	"ai_eng_11_ai_system_design":                       "ai_engineer_m11_ai_system_design_and_architecture",
	"ai_eng_12_ai_governance_and_strategy":             "ai_engineer_m12_ai_governance_ethics_and_strategic_delivery",
	"arch_01_networking_fundamentals":                  "software_architect_m01_networking_foundations_and_protocol_design",
	"arch_02_oop_solid_principles":                     "software_architect_m01_networking_foundations_and_protocol_design",
	"arch_03_relational_data_modeling":                 "software_architect_m02_relational_data_modeling_and_acid_compliance",
	"arch_04_api_design_and_contracts":                 "software_architect_m03_api_design_contracts_and_integration_patterns",
	"arch_05_distributed_data_systems":                 "software_architect_m04_distributed_data_systems_and_cap_trade_offs",
	"arch_06_event_driven_messaging":                   "software_architect_m05_event_driven_architecture_and_messaging_infrastructure",
	"arch_07_data_caching_and_scalability":             "software_architect_m06_caching_performance_tuning_and_gateways",
	"arch_08_microservices_and_cqrs":                   "software_architect_m07_microservices_decomposition_cqrs_and_event_sourcing",
	"arch_09_containerization_and_kubernetes":          "software_architect_m08_containerization_kubernetes_and_cloud_runtimes",
	"arch_10_application_security_and_identity":        "software_architect_m09_application_security_cryptography_and_identity",
	"arch_11_system_resilience_and_observability":      "software_architect_m10_system_resilience_reliability_and_observability",
	"arch_12_enterprise_architecture_and_governance":   "software_architect_m11_enterprise_architecture_governance_and_scaled_delivery",
	"be_01_programming_fundamentals":                   "backend_engineer_m01_programming_fundamentals_and_clean_code",
	"be_02_data_structures_and_algorithms":             "backend_engineer_m02_data_structures_and_algorithms_dsa",
	"be_03_oop_and_design_patterns":                    "backend_engineer_m03_object_oriented_design_and_design_patterns",
	"be_04_database_design_and_sql":                    "backend_engineer_m04_databases_sql_and_relational_modeling",
	"be_05_backend_frameworks_and_apis":                "backend_engineer_m03_object_oriented_design_and_design_patterns",
	"be_06_system_design_and_scalability":              "backend_engineer_m05_distributed_systems_and_scalability",
	"be_07_asynchronous_processing_and_messaging":      "backend_engineer_m06_async_processing_and_message_driven_architecture",
	"be_08_security_auth_and_identity":                 "backend_engineer_m07_security_authentication_and_identity",
	"be_09_containerization_and_infrastructure":         "backend_engineer_m08_containers_deployment_and_infrastructure",
	"be_10_resilience_and_observability":               "backend_engineer_m09_reliability_resilience_and_observability",
	"be_11_microservices_and_architecture_patterns":   "backend_engineer_m10_microservices_and_backend_architecture",
	"be_12_backend_strategy_and_technical_leadership":  "backend_engineer_m11_backend_strategy_governance_and_technical_leadership",
	"da_01_analytics_foundations":                      "data_analyst_m08_machine_learning_foundations_for_analysts",
	"da_02_excel_analysis_and_reporting":               "data_analyst_m06_business_reporting_and_dashboarding",
	"da_03_sql_and_data_access":                        "data_analyst_m02_sql_for_data_retrieval_and_transformation",
	"da_04_python_for_analysis":                        "data_analyst_m03_python_and_data_manipulation_libraries",
	"da_05_data_collection_and_cleaning":               "data_analyst_m02_sql_for_data_retrieval_and_transformation",
	"da_06_descriptive_statistics":                     "data_analyst_m04_descriptive_statistics_and_data_exploration_eda",
	"da_07_data_visualization":                         "data_analyst_m05_data_visualization_and_storytelling",
	"da_08_reporting_and_dashboards":                   "data_analyst_m06_business_reporting_and_dashboarding",
	"da_09_hypothesis_testing_and_regression":          "data_analyst_m07_hypothesis_testing_correlation_and_validation",
	"da_10_machine_learning_basics":                    "data_analyst_m08_machine_learning_foundations_for_analysts",
	"da_11_big_data_and_data_platforms":                "data_analyst_m09_big_data_concepts_and_data_platforms",
	"da_12_portfolio_and_community":                    "data_analyst_m10_portfolio_practice_kaggle_and_continuous_analytics_delivery",
	"ds_01_devops_foundations":                         "devops_sre_m01_devops_foundations_and_operating_systems",
	"ds_02_linux_and_shell":                            "devops_sre_m01_devops_foundations_and_operating_systems",
	"ds_03_version_control_and_ci_cd":                  "devops_sre_m02_version_control_git_and_ci_cd_pipelines",
	"ds_04_networking_security_and_proxies":            "devops_sre_m04_configuration_management_and_secret_handling",
	"ds_05_containers_and_orchestration":               "devops_sre_m03_cloud_platforms_iac_and_provisioning",
	"ds_06_cloud_infrastructure_and_iaac":              "devops_sre_m03_cloud_platforms_iac_and_provisioning",
	"ds_07_configuration_management_and_secrets":      "devops_sre_m04_configuration_management_and_secret_handling",
	"ds_08_monitoring_observability_and_logging":       "devops_sre_m05_monitoring_observability_and_log_management",
	"ds_09_artifact_management_and_gitops":             "devops_sre_m06_artifact_management_delivery_and_containers",
	"ds_10_service_mesh_and_resilience":                "devops_sre_m07_service_mesh_traffic_control_and_resilience",
	"ds_11_incident_response_and_sre_practice":         "devops_sre_m08_incident_response_reliability_and_sre",
	"ds_12_platform_excellence_and_career_growth":      "devops_sre_m09_platform_excellence_automation_and_developer_experience_idps",
	"fe_01_internet_and_web_foundations":               "frontend_engineer_m01_internet_http_and_web_foundations",
	"fe_02_html_css_foundations":                       "frontend_engineer_m02_html_css_and_semantic_layout_foundations",
	"fe_03_javascript_basics":                          "frontend_engineer_m03_javascript_fundamentals_and_dom_manipulation",
	"fe_04_version_control_and_tooling":                "frontend_engineer_m04_version_control_package_managers_and_tooling",
	"fe_05_css_architecture_and_design_systems":        "frontend_engineer_m02_html_css_and_semantic_layout_foundations",
	"fe_06_typescript_and_toolchain":                   "frontend_engineer_m05_typescript_tooling_and_frontend_build_systems",
	"fe_07_frontend_frameworks":                        "frontend_engineer_m06_frontend_frameworks_and_component_design",
	"fe_08_testing_and_quality":                        "frontend_engineer_m07_testing_quality_gates_and_frontend_ci_cd",
	"fe_09_web_security_and_auth":                      "frontend_engineer_m08_web_security_auth_and_browser_safety",
	"fe_10_performance_and_browser_apis":               "frontend_engineer_m09_performance_web_apis_and_browser_internals",
	"fe_11_ssr_ssg_and_fullstack_patterns":             "frontend_engineer_m10_ssr_ssg_and_modern_frontend_delivery_patterns",
	"fe_12_frontend_strategy_and_product_craft":        "frontend_engineer_m11_frontend_strategy_product_craft_and_system_scalability",
	"fs_01_web_foundations":                            "full_stack_m01_web_foundations_and_http_basics",
	"fs_02_git_and_tooling":                            "full_stack_m02_git_github_and_project_tooling",
	"fs_03_html_css_javascript":                        "full_stack_m03_react_tailwind_css_and_component_ui",
	"fs_04_frontend_interactivity":                     "full_stack_m03_react_tailwind_css_and_component_ui",
	"fs_05_react_tailwind":                             "full_stack_m03_react_tailwind_css_and_component_ui",
	"fs_06_backend_foundations":                        "full_stack_m04_backend_foundations_with_node_js",
	"fs_07_restful_apis":                               "full_stack_m05_restful_apis_and_crud_applications",
	"fs_08_data_auth_and_storage":                      "full_stack_m06_authentication_redis_and_data_persistence",
	"fs_09_linux_and_cloud_basics":                     "full_stack_m07_linux_basics_and_cloud_fundamentals",
	"fs_10_cicd_and_monitoring":                        "full_stack_m08_ci_cd_monitoring_and_delivery_pipelines",
	"fs_11_infrastructure_automation":                  "full_stack_m09_infrastructure_automation_and_devops_practice",
	"fs_12_fullstack_delivery":                         "full_stack_m10_production_deployment_monitoring_and_scaling",
	"me_01_mobile_platform_and_language":               "mobile_engineer_m01_android_hello_world_gradle_and_build_tools",
	"me_02_kotlin_oop_and_datastructures":              "mobile_engineer_m01_android_hello_world_gradle_and_build_tools",
	"me_03_android_hello_world_and_gradle":             "mobile_engineer_m01_android_hello_world_gradle_and_build_tools",
	"me_04_version_control_and_collaboration":          "mobile_engineer_m02_version_control_and_team_collaboration",
	"me_05_android_ui_and_navigation":                  "mobile_engineer_m03_ui_components_layouts_navigation_and_views",
	"me_06_architecture_patterns_and_state":            "mobile_engineer_m04_architecture_patterns_and_state_management",
	"me_07_persistence_and_networking":                 "mobile_engineer_m05_storage_data_persistence_and_networking",
	"me_08_dependency_injection_and_services":          "mobile_engineer_m06_dependency_injection_and_common_android_libraries",
	"me_09_testing_debugging_and_quality":              "mobile_engineer_m07_testing_debugging_and_quality_gates",
	"me_10_distribution_and_release":                   "mobile_engineer_m08_distribution_app_release_and_store_deployment",
	"me_11_jetpack_compose_and_modern_ui":              "mobile_engineer_m09_jetpack_compose_and_modern_ui_development",
	"me_12_platform_strategy_and_advanced_mobile_engineering": "mobile_engineer_m10_platform_strategy_and_advanced_mobile_engineering",
	"ml_01_programming_fundamentals":                   "machine_learning_m01_programming_fundamentals_for_ml",
	"ml_04_python_data_stack":                          "machine_learning_m02_data_collection_sources_and_preprocessing",
	"ml_05_data_collection_and_preprocessing":          "machine_learning_m02_data_collection_sources_and_preprocessing",
	"ml_06_machine_learning_basics":                    "machine_learning_m03_machine_learning_fundamentals_and_learning",
	"ml_07_supervised_learning":                        "machine_learning_m04_supervised_learning_algorithms",
	"ml_08_unsupervised_and_reinforcement_learning":    "machine_learning_m05_unsupervised_learning_and_reinforcement",
	"ml_09_model_evaluation_and_validation":            "machine_learning_m06_model_evaluation_validation_and_metrics",
	"ml_10_deep_learning_basics":                       "machine_learning_m07_deep_learning_fundamentals",
	"ml_11_cnn_rnn_transformers":                       "machine_learning_m08_cnns_rnns_and_transformer_architectures",
	"ml_12_mlops_and_production":                       "machine_learning_m09_mlops_deployment_and_production_ml",
	"pm_01_product_foundations":                        "product_manager_m01_product_management_foundations",
	"pm_02_customer_research_and_market_analysis":      "product_manager_m01_product_management_foundations",
	"pm_03_product_strategy_and_positioning":           "product_manager_m01_product_management_foundations",
	"pm_04_prd_user_stories_and_roadmaps":              "product_manager_m02_prds_user_stories_and_roadmaps",
	"pm_05_product_design_and_experimentation":         "product_manager_m03_product_design_ux_and_experimentation",
	"pm_06_agile_and_execution":                        "product_manager_m04_agile_delivery_and_product_execution",
	"pm_07_growth_metrics_and_decision_making":         "product_manager_m05_growth_metrics_analytics_and_data_driven_product_management",
	"pm_08_stakeholder_management_and_communication":   "product_manager_m06_stakeholder_management_and_communication",
	"pm_09_risk_management_and_tooling":                "product_manager_m07_risk_management_tools_and_prioritization_frameworks",
	"pm_10_go_to_market_and_launch":                    "product_manager_m08_go_to_market_strategy_and_launch_planning",
	"pm_11_leadership_and_scaling":                     "product_manager_m09_leadership_influence_and_scaling_products",
	"pm_12_continuous_learning_and_advanced_topics":    "product_manager_m10_continuous_learning_and_advanced_product_strategy",
}

func normalizeRoleID(roleID string) string {
	r := strings.TrimSpace(strings.ToLower(roleID))
	if mapped, ok := domainUUIDMap[r]; ok {
		return mapped
	}
	if r == "software_architecture" {
		return "software_architect"
	}
	return r
}

func (g *GoldResourceLookup) GetRole(roleID string) (GoldRole, bool) {
	roleID = strings.TrimSpace(strings.ToLower(roleID))
	if roleID == "data_engineer" {
		return GoldRole{}, false
	}
	roleID = normalizeRoleID(roleID)
	role, ok := g.Roles[roleID]
	return role, ok
}

// GetGoldModuleResources returns gold tier resources for roleID + moduleQuery (ID, number, or title)
func (g *GoldResourceLookup) GetGoldModuleResources(roleID, moduleQuery string) (*GoldModule, bool) {
	rawRole := strings.TrimSpace(strings.ToLower(roleID))
	if rawRole == "data_engineer" {
		return nil, false
	}
	roleID = normalizeRoleID(roleID)
	role, ok := g.Roles[roleID]
	if !ok {
		return nil, false
	}

	moduleQueryLower := strings.ToLower(strings.TrimSpace(moduleQuery))

	// 1. Check Node Alias Map for graph node IDs
	if targetGoldID, exists := nodeAliasMap[moduleQueryLower]; exists {
		for _, m := range role.Modules {
			if strings.EqualFold(m.ModuleID, targetGoldID) {
				return &m, true
			}
		}
	}

	// 2. Exact ID or Name match
	for _, m := range role.Modules {
		if strings.EqualFold(m.ModuleID, moduleQueryLower) || strings.EqualFold(m.ModuleName, moduleQueryLower) {
			return &m, true
		}
	}

	// 3. Keyword / Title Overlap Match
	var bestMatch *GoldModule
	bestOverlap := 0
	queryWords := tokenizeQuery(moduleQueryLower)

	for _, m := range role.Modules {
		mNameLower := strings.ToLower(m.ModuleName)
		mIDLower := strings.ToLower(m.ModuleID)

		// Substring check
		if strings.Contains(mIDLower, moduleQueryLower) || strings.Contains(mNameLower, moduleQueryLower) || strings.Contains(moduleQueryLower, mNameLower) {
			return &m, true
		}

		// Token overlap check
		mWords := tokenizeQuery(mNameLower)
		overlap := 0
		for w := range queryWords {
			if mWords[w] {
				overlap++
			}
		}
		if overlap > bestOverlap {
			bestOverlap = overlap
			bestMatch = &m
		}
	}

	if bestMatch != nil && bestOverlap >= 2 {
		return bestMatch, true
	}

	// 4. Exact ModuleNumber match (e.g. query is "1" or "4")
	var modNum int
	if _, err := fmt.Sscanf(moduleQueryLower, "%d", &modNum); err == nil && modNum > 0 {
		for _, m := range role.Modules {
			if m.ModuleNumber == modNum {
				return &m, true
			}
		}
	}

	// 5. Fallback to 1st module if query is '1' or '01' and role has modules
	if (moduleQueryLower == "1" || moduleQueryLower == "01") && len(role.Modules) > 0 {
		return &role.Modules[0], true
	}

	if bestMatch != nil {
		return bestMatch, true
	}

	return nil, false
}

func tokenizeQuery(text string) map[string]bool {
	words := make(map[string]bool)
	for _, w := range strings.Fields(strings.ToLower(text)) {
		w = strings.Trim(w, ",.()[]{}!?:;\"'")
		if len(w) > 2 && w != "and" && w != "the" && w != "for" && w != "with" && w != "from" && w != "into" {
			words[w] = true
		}
	}
	return words
}
