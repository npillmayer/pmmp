/*
Package variables implements variables for programming languages similar
to those in MetaFont and MetaPost.

Variables are complex things in MetaFont/MetaPost. These are legal:

   metafont> showvariable x;
   x=1
   x[]=numeric
   x[][]=numeric
   x[][][]=numeric
   x[][][][]=numeric
   x[]r=numeric
   x[]r[]=numeric
   ...

Identifier-strings are called "tags". In the example above, 'x' is a tag
and 'r' is a suffix.

Array variables may be referenced without brackets, if the subscript is just
a numeric literal, i.e. x[2]r and x2r refer to the same variable. We do
not rely on the parser to decipher these kinds of variable names for us,
but rather break up x2r16a => x[2]r[16]a by hand. However, the parser will
split up array indices in brackets, for the subscript may be a complex expression
("x[ypart ((8,5) rotated 20)]" is a valid expression in MetaFont).
Things are further complicated by the fact that subscripts are allowed to
be decimals: x[1.2] is valid, and may be typed "x1.2".

   metafont> x[ypart ((8,5) rotated 20)] = 1;
   ## x7.4347=1

I don't know if this makes sense in practice, but let's try to implement it --
it might be fun!

I did reject some of MetaFont's conventions, however, for the sake of simlicity:
Types are inherited from the tag, i.e. if x is of type numeric, then x[2]r is
of type numeric, too. This is different from MetaFont, where x2r may be of a
different type than x2. Nevertheless, I'll stick to my interpretation,
which I find less confusing.

The implementation currently is tightly coupled to the ANTLR V4 parser
generator. Using ANTLR vor this task is a bit of overkill. Maybe I'll
some day write a recursive descent parser from scratch as a substitute.

___________________________________________________________________________

License

Governed by a 3-Clause BSD license. License file may be found in the root
folder of this module.

Copyright © 2017–2021 Norbert Pillmayer <norbert@pillmayer.com>

*/
package variables
