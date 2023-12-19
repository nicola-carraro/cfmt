# CFMT
A formatter for C code.

## Build
    go build

## Usage
    cfmt [-stdout] path1 [path2 ...]
You must provide at least one path. They must all contain valid C. File contents are overwritten
with formatted text.

If you provide the -stdout flag, files are not overwritten, and the formatted text is printed to
standard output.

## Features
Cfmt is "opinionated", as they say. In other words, it supports only one style and is not configurable.

Here is an example of how output looks like:

    // No blank lines between include directives
    #include <stdio.h>
    #include <stdbool.h>
    #include "something.h"
    #include "stdlib.h"
    
    // Otherwise exactly one blank line between preprocessor directives
    #ifndef SOME_DEFINE
    
    #define SOME_DEFINE
    
    // Function-like macros are formatted more or less like functions
    #define SOME_MACRO(a1, a2) {\
        foo(a1, a2);\
        bar(a1, a2);\
    }
    
    struct Baz {
        int baz;
        char *bazz;
    } Baz;

    // Exactly one line between top-level declarations
    struct Foo {
        int baz;
        char *bazz;
        struct {
            int lucy;
            int mary;
        } Toto;
    } Foo;
    
    /*
       Multiline comments are formatted like this,
       with /* on separate lines and indented text.
    */
    struct *Foo foo() {
        int a = b + c; // No blank line at the beginning and end of a block
        printf("%d\n", a); // A single, optional line between statements
        struct *Foo result = malloc(sizeof(Foo));
        return result;
    }
    
    /*
       If we hit the 80 columns threshold when declaring function arguments,
       they go on separate lines
    */
    bool baz(
        int a,
        int b,
        int c,
        char d,
        char e,
        char f,
        char d,
        char e,
        char f
    ) {
        // Same with function calls
        something(
            alpha,
            beta,
            gamma,
            delta,
            epsilon,
            zeta,
            eta,
            iota,
            kappa,
            lamba,
            mu,
            nu
        );

        int anInt = 1;
        double aDouble = 1575e-2;
        long aLong = 1555.0L;
    
        /*
           We allow statements to wrap anywhere, if they hit the 80 columns threshold,
           indenting the continuation lines
        */
        bool result = (anInt >= aDouble)
            && (aLong <= anInt)
            && (aDouble < 3 && anInt++ < ALong);

        return result;
    }
