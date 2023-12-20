# cfmt
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
cfmt is "opinionated", as they say. In other words, it supports only one style and is not configurable.

Here is an example of how output looks like:

    // No blank lines between include directives
    #include <stdio.h>
    #include <stdbool.h>
    #include "something.h"
    #include "stdlib.h"
    
    // Otherwise exactly one blank line between preprocessor directives
    #ifndef SOME_DEFINE
    
    #define SOME_DEFINE
    
    /*
       We do not insert new lines after { and before } in macros:
       not all function-like macros want to be formatted like functions
    */
    #define MAKE_STRING(s) {.text = s, .len = sizeof(s)}
    
    /*
       If you want to format a macro like a function, you have to do so explictly,
       by adding escaped lines where appropriate
    */
    #define SOME_MACRO(a1, a2) {\
        foo(a1, a2);\
        bar(a1, a2);\
    }
    
    typedef enum Size {
        BIG,
        MID,
        SMALL
    } Size;
    
    // Exactly one line between top level declarations
    struct Baz {
        int baz;
        char *bazz;
    } Baz;
    
    struct Foo {
        int baz;
        char *bazz;
        struct {
            int lucy;
            int mary;
        } Toto;
    } Foo;
    
    /*
       We allow initialiser lists to wrap anywhere, if they hit the 80 columns threshold
    */
    const u8 FONT_SPRITES[] = {
        0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
        0x20, 0x60, 0x20, 0x20, 0x70, // 1
        0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
        0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
        0x90, 0x90, 0xF0, 0x10, 0x10, // 4
        0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
        0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
        0xF0, 0x10, 0x20, 0x40, 0x40, // 7
        0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
        0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
        0xF0, 0x90, 0xF0, 0x90, 0x90, // A
        0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
        0xF0, 0x80, 0x80, 0x80, 0xF0, // C
        0xE0, 0x90, 0x90, 0x90, 0xE0, // D
        0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
        0xF0, 0x80, 0xF0, 0x80, 0x80, // F
    };
    
    /*
       Multiline comments are formatted like this,
       with /* on separate lines and indented text
    */
    struct *Foo foo() {
        // No blank line at the beginning and end of a block
        enum Color {
            RED,
            GREEN,
            BLUE
        };
    
        int a = b + c;
    
        // A single, optional line between statements
        printf("%d\n", a);
    
        struct *Foo result = malloc(sizeof(Foo)) return result;
    }
    
    /*
       If we hit the 80 columns threshold when declaring function arguments,
       they go on separate lines
    */
    struct *Baz baz(
        int a,
        int b,
        int c,
        char d,
        char e,
        char f,
        char d char e,
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
    }


