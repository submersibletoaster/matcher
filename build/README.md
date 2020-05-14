# Building from unscii-16.bdf

from unscii-src ; 
`hex2bdf < unscii-16.hex > unscii-16.bdf`

from pixfont ;
`bdf2pixfont unscii-16.pdf > unscii-16.pixfont`

At this point manual removal of control chars seems necessary to avoid
a bug in fontgen later on. 

Strip everything which is not 8x16; eg A..[12345667]\n
`./strip.pl 14 < unscii-16.pixfont > unscii-8x16.pixfont`

`fontgen -txt unscii-8x16.pixfont -w 8 -h 16 -o unscii`

