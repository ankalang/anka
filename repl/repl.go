package repl

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/ankalang/anka/evaluator"
	"github.com/ankalang/anka/lexer"
	"github.com/ankalang/anka/object"
	"github.com/ankalang/anka/parser"
	"github.com/ankalang/anka/util"
	"github.com/c-bata/go-prompt"

	"golang.org/x/crypto/ssh/terminal"
)


var env *object.Environment


func init() {
	d, _ := os.Getwd()
	env = object.NewEnvironment(os.Stdout, d, "")
}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{}

	for _, key := range env.GetKeys() {
		s = append(s, prompt.Suggest{Text: key})
	}

	if len(d.GetWordBeforeCursor()) == 0 {
		return nil
	}

	return prompt.FilterContains(s, d.GetWordBeforeCursor(), true)
}

var LivePrefixState struct {
	LivePrefix string
	IsEnable   bool
}

func changeLivePrefix() (string, bool) {
	livePrefix := formatLivePrefix(LivePrefixState.LivePrefix)
	return livePrefix, LivePrefixState.IsEnable
}

func Start(in io.Reader, out io.Writer) {
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Println("ANKA interaktif hali başlatılamadı (terminal yok)")
		os.Exit(1)
	}

	promptPrefix := util.GetEnvVar(env, "ANK_PROMPT_PREFIX", ANK_PROMPT_PREFIX)
	livePrompt := util.GetEnvVar(env, "ANK_PROMPT_LIVE_PREFIX", "false")
	if livePrompt == "true" {
		LivePrefixState.LivePrefix = promptPrefix
		LivePrefixState.IsEnable = true
	} else {
		if promptPrefix != formatLivePrefix(promptPrefix) {
			promptPrefix = ANK_PROMPT_PREFIX
		}
	}

	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(promptPrefix),
		prompt.OptionLivePrefix(changeLivePrefix),
		prompt.OptionTitle("anka-repl"),
	)

	p.Run()
}

func executor(line string) {
	if line == "çık" {
		fmt.Printf("%s\n", "Görüşürüz!")
		os.Exit(0)
	}

	if line == "yardım" {
		fmt.Println("Sitemize gel: https://github.com/ankalang/anka")
		return
	}

	Run(line, true)
}

func Run(code string, interactive bool) {
	lex := lexer.New(code)
	p := parser.New(lex)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printParserErrors(p.Errors())
		if !interactive {
			os.Exit(99)
		}
		return
	}

	evaluated := evaluator.BeginEval(program, env, lex)

	if evaluated != nil {
		isError := evaluated.Type() == object.ERROR_OBJ

		if isError {
			fmt.Printf("%s", evaluated.Inspect())
			fmt.Println("")

			if !interactive {
				os.Exit(99)
			}
			return
		}

		if interactive && evaluated.Type() != object.NULL_OBJ {
			fmt.Printf("%s", evaluated.Inspect())
			fmt.Println("")
		}
	}
}

func printParserErrors(errors []string) {
	fmt.Printf("%s", " Ayrıştırıcı hatası:\n")
	for _, msg := range errors {
		fmt.Printf("%s", "\t"+msg+"\n")
	}
}

func BeginRepl(args []string, version string) {
	var interactive bool
	if len(args) == 1 || strings.HasPrefix(args[1], "-") {
		interactive = true
		env.Set("ANK_INTERACTIVE", evaluator.TRUE)
	} else {
		interactive = false
		env.Set("ANK_INTERACTIVE", evaluator.FALSE)
		env.Dir = filepath.Dir(args[1])
	}
	env.Version = version
	env.Set("ANK_VERSION", &object.String{Value: version})
	getAbsInitFile(interactive)

	if interactive {
		for k, v := range evaluator.Fns {
			env.Set(k, v)
		}
		fmt.Printf("Merhaba, Anka programlama diline hoşgeldin! (SÜRÜM: \x1B[38;2;0;200;240m%s\x1B[38;2;255;255;255m)\n", version)

		if r, e := rand.Int(rand.Reader, big.NewInt(100)); e == nil && r.Int64() < 10 {
			if newver, update := util.UpdateAvailable(version); update {
				fmt.Printf("*** Güncelleme mevcut: %s (senin sürümün %s) ***\n", newver, version)
			}
		}
		fmt.Printf("İşin bittiğinde '\x1B[38;2;0;200;240mçık\x1B[38;2;255;255;255m' yaz, kaybolduğunda ise '\x1B[38;2;0;200;240myardım\x1B[38;2;255;255;255m'!\n")
		Start(os.Stdin, os.Stdout)
	} else {

		code, err := ioutil.ReadFile(args[1])
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(99)
		}

		Run(string(code), false)
	}

}
