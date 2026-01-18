package terminal

import (
	"strings"

	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// Banner returns a colored ASCII art banner for Osmedeus
func Banner() string {
	version := HiWhite(strings.Trim(core.VERSION, "-beta"))
	author := Magenta(core.AUTHOR)

	b := Green(`
                                        .;1tfLCL1,
                                       .,,..;i;f0G;
                                             ,:,tCC.  ...
                                             ;i:fCL,1LLtf1i;,
                                           .,::tCL1LC1::;, .,,
                                           ;1:tCL,tLt,1:
                                          ,::tLf, 1Lf;::.
                                        .ii:tLt.  .1Lf;i1.
                                        ,:;tf1      1ft;::
                                     .1;:tf1 `) + HiWhite(` ,i1t1, `) + Green(` ift;;1,
                                    ,i:t;f.`) + HiWhite(` ,LLffLL: `) + Green(` tft;i:
                                    .;:;fff `) + HiWhite(` .LCLLLf,`) + Green(` 1ffi:;.
                                    :fi;Lff1.  `) + HiWhite(`,;;:`) + Green(`  ifffi;f;
                                     .:::tCLLfi:,,:ifLfLt::;.
                                      ,11:1CCCCCLLLLLLf1;1t:
                                      .it;:;1fLLLLfft1;:;ti.
                                         ,:;::;;;;;;;;;;,
                                           .,::::::::,.
`)
	b += "\n" + Green("                            Osmedeus Next Generation ") + version + Green(" by ") + author
	b += "\n\n" + HiCyan("                            "+core.DESC) + "\n"
	b += "\n" + HiWhite(`                                            ¯\_(ツ)_/¯`) + "\n\n"

	return b
}
