package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupDataScrubber(t *testing.T) *DataScrubber {
	customSensitiveWords := []string{
		"consul_token",
		"dd_password",
		"blocked_from_yaml",
	}

	expectedPatterns := make([]string, 0, len(defaultSensitiveWords)+len(customSensitiveWords))
	for _, word := range append(defaultSensitiveWords, customSensitiveWords...) {
		expectedPatterns = append(expectedPatterns, "(?P<key>( +|-)(?i)"+word+")(?P<delimiter> +|=)(?P<value>[^\\s]*)")
	}

	scrubber := NewDefaultDataScrubber()
	scrubber.AddCustomSensitiveWords(customSensitiveWords)

	assert.Equal(t, true, scrubber.Enabled)
	for i, pattern := range scrubber.SensitivePatterns {
		assert.Equal(t, expectedPatterns[i], fmt.Sprint(pattern))
	}

	return scrubber
}

func setupDataScrubberWildCard(t *testing.T) *DataScrubber {
	wildcards := []string{
		"*befpass",
		"afterpass*",
		"*both*",
		"mi*le",
		"*pass*d*",
	}

	expectedPatterns := make([]string, 0, len(defaultSensitiveWords))
	for _, word := range append(defaultSensitiveWords) {
		expectedPatterns = append(expectedPatterns, "(?P<key>( +|-)(?i)"+word+")(?P<delimiter> +|=)(?P<value>[^\\s]*)")
	}

	expectedPatterns = append(expectedPatterns, "(?P<key>( +|-)(?i)[^b]*befpass)(?P<delimiter> +|=)(?P<value>[^\\s]*)")
	expectedPatterns = append(expectedPatterns, "(?P<key>( +|-)(?i)afterpass[^ =]*)(?P<delimiter> +|=)(?P<value>[^\\s]*)")
	expectedPatterns = append(expectedPatterns, "(?P<key>( +|-)(?i)[^b]*both[^ =]*)(?P<delimiter> +|=)(?P<value>[^\\s]*)")
	expectedPatterns = append(expectedPatterns, "(?P<key>( +|-)(?i)mi[^l]*le)(?P<delimiter> +|=)(?P<value>[^\\s]*)")
	expectedPatterns = append(expectedPatterns, "(?P<key>( +|-)(?i)[^p]*pass[^d]*d[^ =]*)(?P<delimiter> +|=)(?P<value>[^\\s]*)")

	scrubber := NewDefaultDataScrubber()
	scrubber.AddCustomSensitiveWords(wildcards)

	assert.Equal(t, true, scrubber.Enabled)
	for i, pattern := range scrubber.SensitivePatterns {
		assert.Equal(t, expectedPatterns[i], fmt.Sprint(pattern))
	}

	return scrubber
}

func setupDataScrubberMatchAll(t *testing.T) *DataScrubber {
	matchAll := []string{
		"*",
	}

	expectedPatterns := make([]string, 0, len(defaultSensitiveWords))
	for _, word := range append(defaultSensitiveWords) {
		expectedPatterns = append(expectedPatterns, "(?P<key>( +|-)(?i)"+word+")(?P<delimiter> +|=)(?P<value>[^\\s]*)")
	}

	expectedPatterns = append(expectedPatterns, "(?P<key>( +|-)(?i)[^ =]*)(?P<delimiter> +|=)(?P<value>[^\\s]*)")

	scrubber := NewDefaultDataScrubber()
	scrubber.AddCustomSensitiveWords(matchAll)

	assert.Equal(t, true, scrubber.Enabled)
	for i, pattern := range scrubber.SensitivePatterns {
		assert.Equal(t, expectedPatterns[i], fmt.Sprint(pattern))
	}

	return scrubber
}

func TestUncompilableWord(t *testing.T) {
	customSensitiveWords := []string{
		"consul_token",
		"dd_password",
		"(an_error",
		")a*",
		"[forbidden]",
		"]a*",
		"blocked_from_yaml",
		"*bef",
		"**bef",
		"after*",
		"after**",
		"*both*",
		"**both**",
		"mi*le",
		"mi**le",
		"*",
		"**",
		"*pass*d*",
	}

	validCustomSenstiveWords := []string{
		"consul_token",
		"dd_password",
		"blocked_from_yaml",
	}

	validWildCards := []string{
		"*bef",
		"after*",
		"*both*",
		"mi*le",
		"*",
		"*pass*d*",
	}

	expectedPatterns := make([]string, 0, len(defaultSensitiveWords)+len(validCustomSenstiveWords))
	for _, word := range append(defaultSensitiveWords, validCustomSenstiveWords...) {
		expectedPatterns = append(expectedPatterns, "(?P<key>( +|-)(?i)"+word+")(?P<delimiter> +|=)(?P<value>[^\\s]*)")
	}

	expectedPatterns = append(expectedPatterns, "(?P<key>( +|-)(?i)[^b]*bef)(?P<delimiter> +|=)(?P<value>[^\\s]*)")
	expectedPatterns = append(expectedPatterns, "(?P<key>( +|-)(?i)after[^ =]*)(?P<delimiter> +|=)(?P<value>[^\\s]*)")
	expectedPatterns = append(expectedPatterns, "(?P<key>( +|-)(?i)[^b]*both[^ =]*)(?P<delimiter> +|=)(?P<value>[^\\s]*)")
	expectedPatterns = append(expectedPatterns, "(?P<key>( +|-)(?i)mi[^l]*le)(?P<delimiter> +|=)(?P<value>[^\\s]*)")
	expectedPatterns = append(expectedPatterns, "(?P<key>( +|-)(?i)[^ =]*)(?P<delimiter> +|=)(?P<value>[^\\s]*)")
	expectedPatterns = append(expectedPatterns, "(?P<key>( +|-)(?i)[^p]*pass[^d]*d[^ =]*)(?P<delimiter> +|=)(?P<value>[^\\s]*)")

	scrubber := NewDefaultDataScrubber()
	scrubber.AddCustomSensitiveWords(customSensitiveWords)

	assert.Equal(t, len(defaultSensitiveWords)+len(validCustomSenstiveWords)+len(validWildCards), len(scrubber.SensitivePatterns))

	assert.Equal(t, true, scrubber.Enabled)
	for i, pattern := range scrubber.SensitivePatterns {
		assert.Equal(t, expectedPatterns[i], fmt.Sprint(pattern))
	}
}

func TestBlacklistedArgs(t *testing.T) {
	cases := []struct {
		cmdline       []string
		parsedCmdline []string
	}{
		{[]string{"agent", "-password", "1234"}, []string{"agent", "-password", "********"}},
		{[]string{"agent", "--password", "1234"}, []string{"agent", "--password", "********"}},
		{[]string{"agent", "-password=1234"}, []string{"agent", "-password=********"}},
		{[]string{"agent", "--password=1234"}, []string{"agent", "--password=********"}},
		{[]string{"fitz", "-consul_token=1234567890"}, []string{"fitz", "-consul_token=********"}},
		{[]string{"fitz", "--consul_token=1234567890"}, []string{"fitz", "--consul_token=********"}},
		{[]string{"fitz", "-consul_token", "1234567890"}, []string{"fitz", "-consul_token", "********"}},
		{[]string{"fitz", "--consul_token", "1234567890"}, []string{"fitz", "--consul_token", "********"}},
		{[]string{"python ~/test/run.py --password=1234 -password 1234 -open_password=admin -consul_token 2345 -blocked_from_yaml=1234 &"},
			[]string{"python", "~/test/run.py", "--password=********", "-password", "********", "-open_password=admin", "-consul_token", "********", "-blocked_from_yaml=********", "&"}},
		{[]string{"agent", "-PASSWORD", "1234"}, []string{"agent", "-PASSWORD", "********"}},
		{[]string{"agent", "--PASSword", "1234"}, []string{"agent", "--PASSword", "********"}},
		{[]string{"agent", "--PaSsWoRd=1234"}, []string{"agent", "--PaSsWoRd=********"}},
		{[]string{"java -password      1234"}, []string{"java", "-password", "", "", "", "", "", "********"}},
	}

	scrubber := setupDataScrubber(t)

	for i := range cases {
		cases[i].cmdline = scrubber.ScrubCmdline(cases[i].cmdline)
		assert.Equal(t, cases[i].parsedCmdline, cases[i].cmdline)
	}
}

func TestBlacklistedArgsWhenDisabled(t *testing.T) {
	cases := []struct {
		cmdline       []string
		parsedCmdline []string
	}{
		{[]string{"agent", "-password", "1234"}, []string{"agent", "-password", "1234"}},
		{[]string{"agent", "--password", "1234"}, []string{"agent", "--password", "1234"}},
		{[]string{"agent", "-password=1234"}, []string{"agent", "-password=1234"}},
		{[]string{"agent", "--password=1234"}, []string{"agent", "--password=1234"}},
		{[]string{"fitz", "-consul_token=1234567890"}, []string{"fitz", "-consul_token=1234567890"}},
		{[]string{"fitz", "--consul_token=1234567890"}, []string{"fitz", "--consul_token=1234567890"}},
		{[]string{"fitz", "-consul_token", "1234567890"}, []string{"fitz", "-consul_token", "1234567890"}},
		{[]string{"fitz", "--consul_token", "1234567890"}, []string{"fitz", "--consul_token", "1234567890"}},
		{[]string{"python ~/test/run.py --password=1234 -password 1234 -open_password=admin -consul_token 2345 -blocked_from_yaml=1234 &"},
			[]string{"python ~/test/run.py --password=1234 -password 1234 -open_password=admin -consul_token 2345 -blocked_from_yaml=1234 &"}},
		{[]string{"agent", "-PASSWORD", "1234"}, []string{"agent", "-PASSWORD", "1234"}},
		{[]string{"agent", "--PASSword", "1234"}, []string{"agent", "--PASSword", "1234"}},
		{[]string{"agent", "--PaSsWoRd=1234"}, []string{"agent", "--PaSsWoRd=1234"}},
		{[]string{"java -password      1234"}, []string{"java -password      1234"}},
	}

	scrubber := setupDataScrubber(t)
	scrubber.Enabled = false

	for i := range cases {
		cases[i].cmdline = scrubber.ScrubCmdline(cases[i].cmdline)
		assert.Equal(t, cases[i].parsedCmdline, cases[i].cmdline)
	}
}

func TestNoBlacklistedArgs(t *testing.T) {
	cases := []struct {
		cmdline       []string
		parsedCmdline []string
	}{
		{[]string{"spidly", "--debug_port=2043"}, []string{"spidly", "--debug_port=2043"}},
		{[]string{"agent", "start", "-p", "config.cfg"}, []string{"agent", "start", "-p", "config.cfg"}},
		{[]string{"p1", "--openpassword=admin"}, []string{"p1", "--openpassword=admin"}},
		{[]string{"p1", "-openpassword", "admin"}, []string{"p1", "-openpassword", "admin"}},
		{[]string{"java -openpassword 1234"}, []string{"java -openpassword 1234"}},
		{[]string{"java -open_password 1234"}, []string{"java -open_password 1234"}},
		{[]string{"java -passwordOpen 1234"}, []string{"java -passwordOpen 1234"}},
		{[]string{"java -password_open 1234"}, []string{"java -password_open 1234"}},
		{[]string{"java -password1 1234"}, []string{"java -password1 1234"}},
		{[]string{"java -password_1 1234"}, []string{"java -password_1 1234"}},
		{[]string{"java -1password 1234"}, []string{"java -1password 1234"}},
		{[]string{"java -1_password 1234"}, []string{"java -1_password 1234"}},
	}

	scrubber := setupDataScrubber(t)

	for i := range cases {
		cases[i].cmdline = scrubber.ScrubCmdline(cases[i].cmdline)
		assert.Equal(t, cases[i].parsedCmdline, cases[i].cmdline)
	}

}

func TestMatchWildCards(t *testing.T) {
	cases := []struct {
		cmdline       []string
		parsedCmdline []string
	}{
		{[]string{"spidly", "--befpass=2043", "onebefpass", "1234", "--befpassCustom=1234"},
			[]string{"spidly", "--befpass=********", "onebefpass", "********", "--befpassCustom=1234"}},
		{[]string{"spidly --befpass=2043 onebefpass 1234 --befpassCustom=1234"},
			[]string{"spidly", "--befpass=********", "onebefpass", "********", "--befpassCustom=1234"}},
		{[]string{"spidly   --befpass=2043   onebefpass   1234   --befpassCustom=1234"},
			[]string{"spidly", "", "", "--befpass=********", "", "", "onebefpass", "", "", "********", "", "", "--befpassCustom=1234"}},

		{[]string{"spidly", "--afterpass=2043", "afterpass_1", "1234", "--befafterpass_1=1234"},
			[]string{"spidly", "--afterpass=********", "afterpass_1", "********", "--befafterpass_1=1234"}},
		{[]string{"spidly --afterpass=2043 afterpass_1 1234 --befafterpass_1=1234"},
			[]string{"spidly", "--afterpass=********", "afterpass_1", "********", "--befafterpass_1=1234"}},
		{[]string{"spidly   --afterpass=2043   afterpass_1   1234   --befafterpass_1=1234"},
			[]string{"spidly", "", "", "--afterpass=********", "", "", "afterpass_1", "", "", "********", "", "", "--befafterpass_1=1234"}},

		{[]string{"spidly", "both", "1234", "-dd_both", "1234", "bothafter", "1234", "--dd_bothafter=1234"},
			[]string{"spidly", "both", "********", "-dd_both", "********", "bothafter", "********", "--dd_bothafter=********"}},
		{[]string{"spidly both 1234 -dd_both 1234 bothafter 1234 --dd_bothafter=1234"},
			[]string{"spidly", "both", "********", "-dd_both", "********", "bothafter", "********", "--dd_bothafter=********"}},
		{[]string{"spidly   both   1234   -dd_both   1234   bothafter   1234   --dd_bothafter=1234"},
			[]string{"spidly", "", "", "both", "", "", "********", "", "", "-dd_both", "", "", "********", "", "", "bothafter", "", "", "********", "", "", "--dd_bothafter=********"}},

		{[]string{"spidly", "middle", "1234", "-mile", "1234", "--mill=1234"},
			[]string{"spidly", "middle", "********", "-mile", "********", "--mill=1234"}},
		{[]string{"spidly middle 1234 -mile 1234 --mill=1234"},
			[]string{"spidly", "middle", "********", "-mile", "********", "--mill=1234"}},
		{[]string{"spidly   middle   1234   -mile   1234   --mill=1234"},
			[]string{"spidly", "", "", "middle", "", "", "********", "", "", "-mile", "", "", "********", "", "", "--mill=1234"}},

		{[]string{"spidly", "--passwd=1234", "password", "1234", "-mypassword", "1234", "--passwords=12345,123456", "--mypasswords=1234,123456"},
			[]string{"spidly", "--passwd=********", "password", "********", "-mypassword", "********", "--passwords=********", "--mypasswords=********"}},
		{[]string{"spidly --passwd=1234 password 1234 -mypassword 1234 --passwords=12345,123456 --mypasswords=1234,123456"},
			[]string{"spidly", "--passwd=********", "password", "********", "-mypassword", "********", "--passwords=********", "--mypasswords=********"}},
		{[]string{"spidly   --passwd=1234   password   1234   -mypassword   1234   --passwords=12345,123456   --mypasswords=1234,123456"},
			[]string{"spidly", "", "", "--passwd=********", "", "", "password", "", "", "********", "", "", "-mypassword", "", "", "********",
				"", "", "--passwords=********", "", "", "--mypasswords=********"}},
	}

	scrubber := setupDataScrubberWildCard(t)

	for i := range cases {
		cases[i].cmdline = scrubber.ScrubCmdline(cases[i].cmdline)
		assert.Equal(t, cases[i].parsedCmdline, cases[i].cmdline)
	}
}

func TestMatchAll(t *testing.T) {
	cases := []struct {
		cmdline       []string
		parsedCmdline []string
	}{
		{[]string{"run", "password", "1234", "-admin", "admin", "--key=1234"},
			[]string{"run", "password", "********", "-admin", "********", "--key=********"}},
		{[]string{"run password 1234 -admin admin --key=1234"},
			[]string{"run", "password", "********", "-admin", "********", "--key=********"}},
		{[]string{"run   password   1234   -admin   admin   --key=1234"},
			[]string{"run", "", "", "password", "", "", "********", "", "", "-admin", "", "", "********", "", "", "--key=********"}},
	}

	scrubber := setupDataScrubberMatchAll(t)

	for i := range cases {
		cases[i].cmdline = scrubber.ScrubCmdline(cases[i].cmdline)
		assert.Equal(t, cases[i].parsedCmdline, cases[i].cmdline)
	}
}

func BenchmarkRegexMatching1(b *testing.B)    { benchmarkRegexMatching(1, b) }
func BenchmarkRegexMatching10(b *testing.B)   { benchmarkRegexMatching(10, b) }
func BenchmarkRegexMatching100(b *testing.B)  { benchmarkRegexMatching(100, b) }
func BenchmarkRegexMatching1000(b *testing.B) { benchmarkRegexMatching(1000, b) }

var avoidOptimization []string

func benchmarkRegexMatching(nbProcesses int, b *testing.B) {
	runningProcesses := make([][]string, nbProcesses)
	foolCmdline := []string{"python ~/test/run.py --password=1234 -password 1234 -password=admin -open_password 2345 -consul=1234 -p 2808 &"}

	customSensitiveWords := []string{
		"consul_token",
		"dd_password",
		"blocked_from_yaml",
	}
	scrubber := NewDefaultDataScrubber()
	scrubber.AddCustomSensitiveWords(customSensitiveWords)

	for i := 0; i < nbProcesses; i++ {
		runningProcesses = append(runningProcesses, foolCmdline)
	}

	var r []string
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, p := range runningProcesses {
			r = scrubber.ScrubCmdline(p)
		}
	}

	avoidOptimization = r
}
