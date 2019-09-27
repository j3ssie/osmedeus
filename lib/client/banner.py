import os
import sys
sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))

# Console colors
W = '\033[1;0m'   # white
R = '\033[1;31m'  # red
G = '\033[1;32m'  # green
O = '\033[1;33m'  # orange
B = '\033[1;34m'  # blue
Y = '\033[1;93m'  # yellow
P = '\033[1;35m'  # purple
C = '\033[1;36m'  # cyan
GR = '\033[1;37m'  # gray

# Just a banner


def banner_(__version__, __author__):
    print(r"""{1}

                                       `@@`
                                      @@@@@@
                                    .@@`  `@@.
                                    :@      @:
                                    :@  {5}:@{1}  @:                       
                                    :@  {5}:@{1}  @:                       
                                    :@      @:                             
                                    `@@.  .@@`
                                      @@@@@@
                                        @@
                                     {0}@{1}  {1}@@  {0}@{1}               
                                    {0}+@@{1} {1}@@ {0}@@+{1}                    
                                 {5}@@:@#@,{1}{1}@@,{5}@#@:@@{1}           
                                ;@+@@`#@@@@#`@@+@;
                                @+ #@@@@@@@@@@# +@
                               @@  @+`@@@@@@`+@  @@
                               @.  @   ;@@;   @  .@
                              {0}#@{1}  {0}'@{1}          {0}@;{1}  {0}@#{1}


                            Osmedeus v{5}{6}{1} by {2}{7}{1}

                                    ¯\_(ツ)_/¯
        """.format(C, G, P, R, B, GR, __version__, __author__))
