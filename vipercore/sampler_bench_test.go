

package vipercore_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gottingen/viper/internal/vtest"
	. "github.com/gottingen/viper/vipercore"
)

var counterTestCases = [][]string{
	// some stuff I made up
	{
		"foo",
		"bar",
		"baz",
		"alpha",
		"bravo",
		"charlie",
		"delta",
	},

	// shuf -n50 /usr/share/dict/words
	{
		"unbracing",
		"stereotomy",
		"supranervian",
		"moaning",
		"exchangeability",
		"gunyang",
		"sulcation",
		"dariole",
		"archheresy",
		"synchronistically",
		"clips",
		"unsanctioned",
		"Argoan",
		"liparomphalus",
		"layship",
		"Fregatae",
		"microzoology",
		"glaciaria",
		"Frugivora",
		"patterist",
		"Grossulariaceae",
		"lithotint",
		"bargander",
		"opisthographical",
		"cacography",
		"chalkstone",
		"nonsubstantialism",
		"sardonicism",
		"calamiform",
		"lodginghouse",
		"predisposedly",
		"topotypic",
		"broideress",
		"outrange",
		"gingivolabial",
		"monoazo",
		"sparlike",
		"concameration",
		"untoothed",
		"Camorrism",
		"reissuer",
		"soap",
		"palaiotype",
		"countercharm",
		"yellowbird",
		"palterly",
		"writinger",
		"boatfalls",
		"tuglike",
		"underbitten",
	},

	// shuf -n100 /usr/share/dict/words
	{
		"rooty",
		"malcultivation",
		"degrade",
		"pseudoindependent",
		"stillatory",
		"antiseptize",
		"protoamphibian",
		"antiar",
		"Esther",
		"pseudelminth",
		"superfluitance",
		"teallite",
		"disunity",
		"spirignathous",
		"vergency",
		"myliobatid",
		"inosic",
		"overabstemious",
		"patriarchally",
		"foreimagine",
		"coetaneity",
		"hemimellitene",
		"hyperspatial",
		"aulophyte",
		"electropoion",
		"antitrope",
		"Amarantus",
		"smaltine",
		"lighthead",
		"syntonically",
		"incubous",
		"versation",
		"cirsophthalmia",
		"Ulidian",
		"homoeography",
		"Velella",
		"Hecatean",
		"serfage",
		"Spermaphyta",
		"palatoplasty",
		"electroextraction",
		"aconite",
		"avirulence",
		"initiator",
		"besmear",
		"unrecognizably",
		"euphoniousness",
		"balbuties",
		"pascuage",
		"quebracho",
		"Yakala",
		"auriform",
		"sevenbark",
		"superorganism",
		"telesterion",
		"ensand",
		"nagaika",
		"anisuria",
		"etching",
		"soundingly",
		"grumpish",
		"drillmaster",
		"perfumed",
		"dealkylate",
		"anthracitiferous",
		"predefiance",
		"sulphoxylate",
		"freeness",
		"untucking",
		"misworshiper",
		"Nestorianize",
		"nonegoistical",
		"construe",
		"upstroke",
		"teated",
		"nasolachrymal",
		"Mastodontidae",
		"gallows",
		"radioluminescent",
		"uncourtierlike",
		"phasmatrope",
		"Clunisian",
		"drainage",
		"sootless",
		"brachyfacial",
		"antiheroism",
		"irreligionize",
		"ked",
		"unfact",
		"nonprofessed",
		"milady",
		"conjecture",
		"Arctomys",
		"guapilla",
		"Sassenach",
		"emmetrope",
		"rosewort",
		"raphidiferous",
		"pooh",
		"Tyndallize",
	},
}

func BenchmarkSampler_Check(b *testing.B) {
	for _, keys := range counterTestCases {
		b.Run(fmt.Sprintf("%v keys", len(keys)), func(b *testing.B) {
			fac := NewSampler(
				NewCore(
					NewJSONEncoder(testEncoderConfig()),
					&vtest.Discarder{},
					DebugLevel,
				),
				time.Millisecond, 1, 1000)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					ent := Entry{
						Level:   DebugLevel + Level(i%4),
						Message: keys[i],
					}
					_ = fac.Check(ent, nil)
					i++
					if n := len(keys); i >= n {
						i -= n
					}
				}
			})
		})
	}
}
