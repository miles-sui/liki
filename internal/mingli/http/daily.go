package minglihttp

import (
	"fmt"
	"time"

	"github.com/25types/25types/internal/ganzhi"
	"github.com/25types/25types/internal/mingli/bazi"
	"github.com/25types/25types/internal/mingli/huangli"
)

// DailySuggestion holds the daily suggestion data.
type DailySuggestion struct {
	Date         string `json:"date"`
	DayPillar    string `json:"day_pillar"`
	JianChu      string `json:"jian_chu"`
	JianChuName  string `json:"jian_chu_name"`
	Suggestion   string `json:"suggestion"`
	Question     string `json:"question"`
	Personalized bool   `json:"personalized"`
}

// jianchuAdvice maps each JianChu god to general daily advice.
var jianchuAdvice = map[string]struct {
	suggestion string
	question   string
}{
	"建": {"今天是建日，适合开启新事物。如果有一直在想但没开始的事，今天是最好的启动日。", "有什么事你一直想开始却没开始的？"},
	"除": {"除日是清理和除旧的日子。适合整理空间、清理思绪、结束拖了很久的事。", "有什么旧习惯或旧情绪你准备好放下了？"},
	"满": {"满日宜祭祀、感恩。盘点你已经拥有的东西，而不是盯着缺失。", "今天你最感恩的三件事是什么？"},
	"平": {"平日能量中和，适合日常维护和稳步推进。不急于做重大决定，专注在已经在做的事上。", "你现在正在进行的事中，哪一件最值得你投入今天的注意力？"},
	"定": {"定日适合签约、承诺、立规矩。今天做出的决定容易落实——但想清楚再定。", "有什么决定你已经想好了，今天该把它定下来了？"},
	"执": {"执日适合深入执行和坚守。你已经开始的计划，今天要坚持做下去，不被干扰。", "在什么地方你最容易分心？今天怎么守住自己的节奏？"},
	"破": {"破日是传统上万事不宜的日子。今天适合休息、反思，不适合做重大决策或启动新事。", "今天不做事也很好——你最近一次真正休息是什么时候？"},
	"危": {"危日能量偏紧，适合安静、安床、内省。今天注意言辞，不要冒不必要的风险。", "在什么情况下你的身体告诉你「不行了」但你忽略了？"},
	"成": {"成日是十二神中最吉之一——宜嫁娶、开业、入宅、签约。今天的能量支持你把事情做成。", "哪件已经推进了一大半的事，今天可以加把劲把它完成？"},
	"收": {"收日宜纳财、收尾、收心。适合做结算、复盘、把散开的能量收回来。", "你最近有什么能量散得太开的地方需要收一收？"},
	"开": {"开日是吉日——宜开业、出行、开启新篇章。今天的能量支持打破边界、扩展视野。", "有什么地方你一直在收缩，今天可以试试打开自己？"},
	"闭": {"闭日宜埋葬、告别、结束。表面上看不吉利，其实是善终的日子——给该结束的事一个体面的句号。", "有哪段关系、哪个阶段、哪种身份，你真的准备好跟它说再见了？"},
}

// ComputeDailySuggestion returns a daily suggestion optionally personalized to the day master.
func ComputeDailySuggestion(dayMaster ganzhi.Stem) DailySuggestion {
	now := time.Now()
	date := now.Format("2006-01-02")
	dp := bazi.DayPillar(now.Year(), int(now.Month()), now.Day())
	jianChu := huangli.LookupJianChu(date)
	dpStr := fmt.Sprintf("%s%s", bazi.DayMasterNameString(dp.Stem), bazi.MonthBranchNameString(dp.Branch))

	s := DailySuggestion{
		Date:        date,
		DayPillar:   dpStr,
		JianChu:     jianChu,
		JianChuName: jianchuName(jianChu),
	}

	if advice, ok := jianchuAdvice[jianChu]; ok {
		s.Suggestion = advice.suggestion
		s.Question = advice.question
	}

	if dayMaster > 0 {
		s.Personalized = true
		dmElement := bazi.StemElement(dayMaster)
		dayElement := bazi.StemElement(dp.Stem)
		rel := huangli.EvaluateGan(dp.Stem, dayMaster)
		s.Suggestion += fmt.Sprintf(" 你的日主%s（%s），今日天干为%s——与你的十神关系是%s。", bazi.DayMasterNameString(dayMaster), ganzhi.Element(dmElement).String(), bazi.DayMasterNameString(dp.Stem), rel)

		if bazi.ElementThatGenerates(dmElement) == dayElement {
			s.Suggestion += " 今天你的能量在向外输出——适合创造和表达，注意不要透支。"
		} else if bazi.ElementThatGenerates(dayElement) == dmElement {
			s.Suggestion += " 今天的能量在滋养你——适合学习和接受，今天的收获会比平时多。"
		} else if bazi.ElementThatControls(dmElement) == dayElement {
			s.Suggestion += " 今天你容易有控制欲——适合做决策和执行，但注意不要强势过头。"
		} else if bazi.ElementThatControls(dayElement) == dmElement {
			s.Suggestion += " 今天的能量对你有些压制——适合低调处理事务，不必强求推进。"
		}
	}

	return s
}

// DayMasterFromBirthInfo computes the day master stem from raw birth parameters.
func DayMasterFromBirthInfo(year, month, day, hour, minute int, longitude, timezone float64) ganzhi.Stem {
	isDST := bazi.IsDST(year, month, day)
	ast := bazi.ComputeSolarTime(year, month, day, hour, minute, longitude, timezone, isDST)
	bz := bazi.ComputeBazi(ast, year, month, day, hour, minute, timezone, isDST)
	return bz.Day.Stem
}

func jianchuName(short string) string {
	names := map[string]string{
		"建": "建日", "除": "除日", "满": "满日", "平": "平日",
		"定": "定日", "执": "执日", "破": "破日", "危": "危日",
		"成": "成日", "收": "收日", "开": "开日", "闭": "闭日",
	}
	if n, ok := names[short]; ok {
		return n
	}
	return short
}
