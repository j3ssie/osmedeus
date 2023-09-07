package core

import (
	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/libs"
)

// Banner print ascii banner
func Banner() string {
	version := color.HiWhiteString(libs.VERSION)
	author := color.MagentaString(libs.AUTHOR)
	//W := color.HiWhiteString(``)
	b := color.GreenString(``)

	b += color.GreenString(`
            
                                        .;1tfLCL1,
                                       .,,..;i;f0G;
                                             ,:,tCC.  ...
                                             ;i:fCL,1LLtf1i;,
                                           .,::tCL1LC1::;, .,,
                                           ;1:tCL,tLt,1:
                                          ,::tLf, 1Lf;::.
                                        .ii:tLt.  .1Lf;i1.
                                        ,:;tf1      1ft;::
                                     .1;:tf1 `) + color.HiWhiteString(` ,i1t1, `) + color.GreenString(` ift;;1,
                                    ,i:t;f.`) + color.HiWhiteString(` ,LLffLL: `) + color.GreenString(` tft;i:
                                    .;:;fff `) + color.HiWhiteString(` .LCLLLf,`) + color.GreenString(` 1ffi:;.
                                    :fi;Lff1.  `) + color.HiWhiteString(`,;;:`) + color.GreenString(`  ifffi;f;
                                     .:::tCLLfi:,,:ifLfLt::;.
                                      ,11:1CCCCCLLLLLLf1;1t:
                                      .it;:;1fLLLLfft1;:;ti.
                                         ,:;::;;;;;;;;;;,
                                           .,::::::::,.
	`)

	//
	//
	//b += "\n\t" + color.GreenString(`                                      @@@@@@`)
	//b += "\n\t" + color.GreenString(`                                    .@@'  '@@.`)
	//b += "\n\t" + color.GreenString(`                                    :@      @:`)
	//b += "\n\t" + color.GreenString(`                                    :@  %v:@`, W) + color.GreenString(`  @:`)
	//b += "\n\t" + color.GreenString(`                                    :@  %v:@`, W) + color.GreenString(`  @:`)
	//
	//b += "\n\t" + color.GreenString(`                                    :@      @:`)
	//b += "\n\t" + color.GreenString(`                                    '@@.  .@@'`)
	//b += "\n\t" + color.GreenString(`                                      @@@@@@`)
	//b += "\n\t" + color.GreenString(`                                        @@`)
	//b += "\n\t" + color.HiCyanString(`                                     @  `) + color.GreenString(`@@`) + color.HiCyanString(`  @`)
	//b += "\n\t" + color.HiWhiteString(`                                    +@@`) + color.GreenString(` @@ `) + color.HiWhiteString(` @@+`)
	//b += "\n\t" + color.GreenString(`                                 @@:@#@,@@,@#@:@@`)
	//b += "\n\t" + color.GreenString(`                                ;@+@@'#@@@@#'@@+@;`)
	//b += "\n\t" + color.GreenString(`                                @+ #@@  @@  @@# +@`)
	//b += "\n\t" + color.GreenString(`                               @@  @+'@@@@@@'+@  @@`)
	//b += "\n\t" + color.GreenString(`                               @.  @   ;@@;   @  .@`)
	//b += "\n\t" + color.BlueString(`                              #@  '@          @;  @#`)

	b += "\n\n\t" + color.GreenString(`                  Osmedeus Next Generation %v`, version) + color.GreenString(` by %v`, author)
	b += "\n\n" + color.HiCyanString(`	                    %s`, libs.DESC) + "\n"
	b += "\n" + color.HiWhiteString(`                                            ¯\_(ツ)_/¯`) + "\n\n"
	color.Unset()
	return b
}
