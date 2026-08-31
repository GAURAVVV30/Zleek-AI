package aiengine

// App holds the shared in-process AI singletons (mirroring the FastAPI module
// globals: GraphEngine(), LLMClient(), get_guardrails_engine(), VoiceEngine()
// are module-level singletons loaded once).

import "sync"

var (
	appOnce     sync.Once
	appInstance *App
	appErr      error
)

type App struct {
	Graph    *GraphEngine
	LLM      *LLMClient
	Voice    *VoiceEngine
	Roadmaps *RoadmapStore
}

// GetApp returns the shared App, initializing the graph engine on first use.
func GetApp() (*App, error) {
	appOnce.Do(func() {
		graph, err := ParseGraphEngine()
		if err != nil {
			appErr = err
			return
		}
		appInstance = &App{
			Graph:    graph,
			LLM:      DefaultLLM(),
			Voice:    NewVoiceEngine(),
			Roadmaps: NewRoadmapStore(),
		}
	})
	return appInstance, appErr
}
