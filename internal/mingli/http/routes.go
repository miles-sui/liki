package minglihttp

import "net/http"

// RegisterRoutes registers all mingli (stateless computation) routes on the given ServeMux.
// Use for the standalone mingli-server. Includes health (no DB), location, and solar-terms.
func RegisterRoutes(mux *http.ServeMux) {
	registerCoreRoutes(mux)

	loh := &LocationHandler{}
	mux.HandleFunc("GET /api/location", loh.GetLocation)
	mux.HandleFunc("GET /api/solar-terms", SolarTerms)
	mux.HandleFunc("GET /api/health", Health)
}

// RegisterCoreRoutes registers only the computation/reference routes (huangli, fengshui,
// reference, career). Safe to call alongside app-server routes without conflicts.
func RegisterCoreRoutes(mux *http.ServeMux) {
	registerCoreRoutes(mux)
}

func registerCoreRoutes(mux *http.ServeMux) {
		RegisterDocs(mux)
	limit := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
			}
			next(w, r)
		}
	}

	hlh := &HuangliHandler{}
	mux.HandleFunc("GET /api/huangli/query", hlh.Query)
	mux.HandleFunc("POST /api/huangli/bond", limit(hlh.Bond))
	mux.HandleFunc("GET /api/huangli/jieqi", hlh.JieQi)


	fsh := &FengShuiHandler{}
	mux.HandleFunc("GET /api/fengshui/san-yuan", fsh.GetSanYuan)
	mux.HandleFunc("POST /api/fengshui/minggua", limit(fsh.MingGua))
	mux.HandleFunc("POST /api/fengshui/hecan", limit(fsh.HeCan))

	refh := &ReferenceHandler{}
	mux.HandleFunc("GET /api/reference/stems", refh.Stems)
	mux.HandleFunc("GET /api/reference/branches", refh.Branches)
	mux.HandleFunc("GET /api/reference/nayin", refh.Nayin)
	mux.HandleFunc("GET /api/reference/shensha", refh.ShenSha)
	mux.HandleFunc("GET /api/reference/zodiac", refh.Zodiac)
	mux.HandleFunc("GET /api/reference/mansions", refh.Mansions)
	mux.HandleFunc("GET /api/reference/trigrams", refh.Trigrams)
	mux.HandleFunc("GET /api/reference/huangdao", refh.HuangDao)
	mux.HandleFunc("GET /api/reference/24-shan", refh.TwentyFourShan)
	mux.HandleFunc("GET /api/reference/cities", refh.Cities)

	bzh := &MingliHandler{}
	mux.HandleFunc("POST /api/bazi/chart", limit(bzh.ComputeChart))
	mux.HandleFunc("POST /api/bazi/bond", limit(bzh.BondCharts))
	mux.HandleFunc("POST /api/bazi/liunian", limit(bzh.LiuNian))
	mux.HandleFunc("POST /api/bazi/liuyue", limit(bzh.LiuYue))
	mux.HandleFunc("POST /api/bazi/liuri", limit(bzh.LiuRi))
	mux.HandleFunc("POST /api/bazi/liushi", limit(bzh.LiuShi))
	mux.HandleFunc("POST /api/bazi/xiao-yun", limit(bzh.XiaoYun))
	mux.HandleFunc("POST /api/bazi/xiao-xian", limit(bzh.XiaoXian))

	qmh := &QimingHandler{}
	mux.HandleFunc("POST /api/qiming/generate", limit(qmh.Generate))
	mux.HandleFunc("POST /api/qiming/evaluate", limit(qmh.Evaluate))
	mux.HandleFunc("GET /api/qiming/characters", qmh.GetCharacters)
}
