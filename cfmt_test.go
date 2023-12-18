package main

import (
	"testing"
	//"fmt"
)

func _testTokenizeSingleToken(t *testing.T, text string, tType TokenType) {

	token := parseToken(text)

	if token.Type != tType {
		t.Errorf("Token should be %s, found %s", tType, token.Type)
	}

	if token.Content != text {
		t.Errorf("Token content should be %s, found %s", text, token.Content)
	}
}

func _testFormat(t *testing.T, input string, expected string) {
	output := format(input)

	for i, r := range []byte(expected) {
		if i >= len(output) {
			t.Errorf("Index %d, expected %s, found end of string", i, output[i:])
			break
		}

		if r != output[i] {
			t.Errorf("Index %d, expected %s, output %s", i, expected[i:], output[i:])
			break
		}
	}

	if output != expected {
		t.Errorf("Output should be:\n%s\nfound:\n%s\n", expected, output)
	}
}

func TestTokenizeString(t *testing.T) {
	_testTokenizeSingleToken(t, "\"toto\"", String)
	_testTokenizeSingleToken(t, "\"to\\\"o\"", String)
	_testTokenizeSingleToken(t, "L\"A wide string\"", String)
}

func TestTokenizeFloat(t *testing.T) {
	_testTokenizeSingleToken(t, "55.0f", Float)
	_testTokenizeSingleToken(t, "123.456e-67", Float)
	_testTokenizeSingleToken(t, "123e+86", Float)
	_testTokenizeSingleToken(t, "15.75", Float)
	_testTokenizeSingleToken(t, "1.575E1", Float)
	_testTokenizeSingleToken(t, "1575e-2", Float)
	_testTokenizeSingleToken(t, "25E-4", Float)
	_testTokenizeSingleToken(t, "10.0L", Float)
	_testTokenizeSingleToken(t, "10.0", Float)
	_testTokenizeSingleToken(t, ".0075e2", Float)
	_testTokenizeSingleToken(t, "0.075e1", Float)
	_testTokenizeSingleToken(t, ".075e1", Float)
	_testTokenizeSingleToken(t, "75e-2", Float)
}

func TestTokenizeIdentifier(t *testing.T) {
	_testTokenizeSingleToken(t, "float_count", Identifier)
}

func TestTokenizeInteger(t *testing.T) {
	_testTokenizeSingleToken(t, "0", Integer)
	_testTokenizeSingleToken(t, "3", Integer)
	_testTokenizeSingleToken(t, "0x1C", Integer)
	_testTokenizeSingleToken(t, "034", Integer)
	_testTokenizeSingleToken(t, "28", Integer)

	_testTokenizeSingleToken(t, "024", Integer)
	_testTokenizeSingleToken(t, "4000000024u", Integer)
	_testTokenizeSingleToken(t, "2000000022l", Integer)
	_testTokenizeSingleToken(t, "4000000000ul", Integer)
	_testTokenizeSingleToken(t, "9000000000LL", Integer)
	_testTokenizeSingleToken(t, "900000000001ull", Integer)
	_testTokenizeSingleToken(t, "9000000000002I64", Integer)
	_testTokenizeSingleToken(t, "90000000000004ui64", Integer)

	_testTokenizeSingleToken(t, "024", Integer)
	_testTokenizeSingleToken(t, "04000000024u", Integer)
	_testTokenizeSingleToken(t, "02000000022l", Integer)
	_testTokenizeSingleToken(t, "04000000000UL", Integer)
	_testTokenizeSingleToken(t, "044000000000000ll", Integer)
	_testTokenizeSingleToken(t, "044400000000000001Ull", Integer)
	_testTokenizeSingleToken(t, "04444000000000000002i64", Integer)
	_testTokenizeSingleToken(t, "04444000000000000004uI64", Integer)

	_testTokenizeSingleToken(t, "0x2a", Integer)
	_testTokenizeSingleToken(t, "0XA0000024u", Integer)
	_testTokenizeSingleToken(t, "0x20000022l", Integer)
	_testTokenizeSingleToken(t, "0XA0000021uL", Integer)
	_testTokenizeSingleToken(t, "0x8a000000000000ll", Integer)
	_testTokenizeSingleToken(t, "0x8A40000000000010uLL", Integer)
	_testTokenizeSingleToken(t, "0x4a44000000000020I64", Integer)
	_testTokenizeSingleToken(t, "0x8a44000000000040Ui64", Integer)

}

func TestTokenizePunctuation(t *testing.T) {
	_testTokenizeSingleToken(t, "+", Punctuation)
	_testTokenizeSingleToken(t, "-", Punctuation)
	_testTokenizeSingleToken(t, "*", Punctuation)
	_testTokenizeSingleToken(t, "/", Punctuation)
	_testTokenizeSingleToken(t, "=", Punctuation)
	_testTokenizeSingleToken(t, "+=", Punctuation)
	_testTokenizeSingleToken(t, "-=", Punctuation)
	_testTokenizeSingleToken(t, "*=", Punctuation)
	_testTokenizeSingleToken(t, "/=", Punctuation)
	_testTokenizeSingleToken(t, "++", Punctuation)
	_testTokenizeSingleToken(t, "--", Punctuation)
	_testTokenizeSingleToken(t, "==", Punctuation)
	_testTokenizeSingleToken(t, "<", Punctuation)
	_testTokenizeSingleToken(t, "<=", Punctuation)
	_testTokenizeSingleToken(t, ">", Punctuation)
	_testTokenizeSingleToken(t, ">=", Punctuation)
	_testTokenizeSingleToken(t, "!=", Punctuation)
	_testTokenizeSingleToken(t, "||", Punctuation)
	_testTokenizeSingleToken(t, "&&", Punctuation)
	_testTokenizeSingleToken(t, "!", Punctuation)
	_testTokenizeSingleToken(t, "&", Punctuation)
	_testTokenizeSingleToken(t, "|", Punctuation)
	_testTokenizeSingleToken(t, "~", Punctuation)
	_testTokenizeSingleToken(t, "^=", Punctuation)
	_testTokenizeSingleToken(t, "&=", Punctuation)
	_testTokenizeSingleToken(t, "|=", Punctuation)
	_testTokenizeSingleToken(t, "^=", Punctuation)
	_testTokenizeSingleToken(t, "<<", Punctuation)
	_testTokenizeSingleToken(t, ">>", Punctuation)
	_testTokenizeSingleToken(t, "<%", Punctuation)
	_testTokenizeSingleToken(t, "%>", Punctuation)
	_testTokenizeSingleToken(t, "<:", Punctuation)
	_testTokenizeSingleToken(t, ":>", Punctuation)
	_testTokenizeSingleToken(t, "%:", Punctuation)
	_testTokenizeSingleToken(t, "%:%:", Punctuation)
	_testTokenizeSingleToken(t, "#", Punctuation)
	_testTokenizeSingleToken(t, "#@", Punctuation)
	_testTokenizeSingleToken(t, "##", Punctuation)
}

func TestFormatStructDecl(t *testing.T) {
	input := `typedef struct {
int bar;     char *baz;}Foo;`

	expected := `typedef struct {
    int bar;
    char *baz;
} Foo;
`

	_testFormat(t, input, expected)

	input = `struct  Foo{
			    int bar;     char *baz;
			}



			;`

	expected = `struct Foo {
    int bar;
    char *baz;
};
`

	_testFormat(t, input, expected)

	input = `struct Foo{

    int bar;

    char *baz;

};
`

	expected = `struct Foo {
    int bar;
    char *baz;
};
`
	_testFormat(t, input, expected)

	input = `typedef struct { struct {C8_Key kp_0;} keypad;} C8_Keypad;`
	expected = `typedef struct {
    struct {
        C8_Key kp_0;
    } keypad;
} C8_Keypad;
`
	_testFormat(t, input, expected)

}

func TestFormatInitializerList(t *testing.T) {
	input := `Foo foo = {
		0    }
		;`
	expected := "Foo foo = {0};\n"

	_testFormat(t, input, expected)

	input = " p = {.x,.y};\n"
	expected = "p = {.x, .y};\n"

	_testFormat(t, input, expected)

	input = " p = {. x, . y};\n"
	expected = "p = {.x, .y};\n"

	_testFormat(t, input, expected)

	input = " p = { .x, .y};\n"
	expected = "p = {.x, .y};\n"

	_testFormat(t, input, expected)

	input = " p = { x, {y,\n z}};\n"
	expected = "p = {\n    x, {\n        y,\n        z\n    }\n};\n"

	_testFormat(t, input, expected)

}

func TestFormatLoop(t *testing.T) {
	input := " for (int i=0;i<3;i++) {printf(\"%d\\n\", i);}"

	expected := "for (int i = 0; i < 3; i++) {\n" +
		"    printf(\"%d\\n\", i);\n" +
		"}\n"

	_testFormat(t, input, expected)

	input = " for (int i=0;\ni<3;\ni++\n) {printf(\"%d\\n\", i);}"

	expected = "for (int i = 0; i < 3; i++) {\n" +
		"    printf(\"%d\\n\", i);\n" +
		"}\n"

	_testFormat(t, input, expected)

	input = "Foo zz = {123, \"123\"  };"
	expected = "Foo zz = {123, \"123\"};\n"
	_testFormat(t, input, expected)

	input = "Foo zz = {\n123,\n\"123\"\n};"
	expected = "Foo zz = {\n    123,\n    \"123\"\n};\n"
	_testFormat(t, input, expected)

	input = "int foo(){for (i16 i = 0; i < buf_size / sizeof(i16); i++) {return bar;}}"
	expected = `int foo() {
    for (i16 i = 0; i < buf_size / sizeof(i16); i++) {
        return bar;
    }
}
`
	_testFormat(t, input, expected)

	input = "{do {foo();} while(bar);}"
	expected = "{\n    do {\n        foo();\n    } while (bar);\n}\n"
	_testFormat(t, input, expected)

}

func TestFormatOperators(t *testing.T) {
	input := "int a=b*c;"
	expected := "int a = b * c;\n"
	_testFormat(t, input, expected)

	input = "aa   -> bar =3;"
	expected = "aa->bar = 3;\n"
	_testFormat(t, input, expected)

	input = "OPENFILENAME file_name = {.hwndOwner = state.window};\n"
	expected = "OPENFILENAME file_name = {.hwndOwner = state.window};\n"
	_testFormat(t, input, expected)

	input = "	OPENFILENAME file_name = {.lStructSize = sizeof(file_name),.hwndOwner = state.window,.hInstance = state.instance,.lpstrFilter = \"Chip 8 rom (*.ch8)\\0*.ch8\\0All files (*.*)'\\0*.*\",.lpstrFile = path,.nMaxFile = C8_ARRCOUNT(path),.lpstrInitialDir = \"roms\"};"
	expected = `OPENFILENAME file_name = {
    .lStructSize = sizeof(file_name),
    .hwndOwner = state.window,
    .hInstance = state.instance,
    .lpstrFilter = \"Chip 8 rom (*.ch8)\\0*.ch8\\0All files (*.*)'\\0*.*\",
    .lpstrFile = path,
    .nMaxFile = C8_ARRCOUNT(path),
    .lpstrInitialDir = \"roms\"
};
`

	input = "a . b = c . d;"
	expected = "a.b = c.d;\n"
	_testFormat(t, input, expected)

	input = "a\n.b = c\n.d;"
	expected = "a.b = c.d;\n"
	_testFormat(t, input, expected)

	input = "i = i ++ == i --;\n"
	expected = "i = i++ == i--;\n"
	_testFormat(t, input, expected)

	input = "i = i\n++ == i\n--;\n"
	expected = "i = i++ == i--;\n"
	_testFormat(t, input, expected)

	input = "i = i++==i--;\n"
	expected = "i = i++ == i--;\n"
	_testFormat(t, input, expected)

	input = "i = i -- == i ++;\n"
	expected = "i = i-- == i++;\n"
	_testFormat(t, input, expected)

	input = "i = i\n-- == i\n++;\n"
	expected = "i = i-- == i++;\n"
	_testFormat(t, input, expected)

	input = "i = i--==i++;\n"
	expected = "i = i-- == i++;\n"
	_testFormat(t, input, expected)

	input = "i = (i) ++ == (i) --;\n"
	expected = "i = (i)++ == (i)--;\n"
	_testFormat(t, input, expected)

	input = "i = (i)\n++ == (i)\n--;\n"
	expected = "i = (i)++ == (i)--;\n"
	_testFormat(t, input, expected)

	input = "i = (i)++== (i)--;\n"
	expected = "i = (i)++ == (i)--;\n"
	_testFormat(t, input, expected)

	input = "i = (i)++ == (i)-- ;\n"
	expected = "i = (i)++ == (i)--;\n"
	_testFormat(t, input, expected)

	input = "i = ((i) ++)-- == ((i) ++)--;\n"
	expected = "i = ((i)++)-- == ((i)++)--;\n"
	_testFormat(t, input, expected)

	input = "i = ((i) --)-- == ((i) --) --;\n"
	expected = "i = ((i)--)-- == ((i)--)--;\n"
	_testFormat(t, input, expected)

	input = "i = (i)++<= (i)--;\n"
	expected = "i = (i)++ <= (i)--;\n"
	_testFormat(t, input, expected)

	input = "i = (i)++ <= (i)-- ;\n"
	expected = "i = (i)++ <= (i)--;\n"
	_testFormat(t, input, expected)

	input = "i = ((i) ++)-- <= ((i) ++)--;\n"
	expected = "i = ((i)++)-- <= ((i)++)--;\n"
	_testFormat(t, input, expected)

	input = "i = ((i) --)-- <= ((i) --) --;\n"
	expected = "i = ((i)--)-- <= ((i)--)--;\n"
	_testFormat(t, input, expected)

	input = "i = (i)++!= (i)--;\n"
	expected = "i = (i)++ != (i)--;\n"
	_testFormat(t, input, expected)

	input = "i = (i)++ != (i)-- ;\n"
	expected = "i = (i)++ != (i)--;\n"
	_testFormat(t, input, expected)

	input = "i = ((i) ++)-- != ((i) ++)--;\n"
	expected = "i = ((i)++)-- != ((i)++)--;\n"
	_testFormat(t, input, expected)

	input = "i = ((i) --)-- != ((i) --) --;\n"
	expected = "i = ((i)--)-- != ((i)--)--;\n"
	_testFormat(t, input, expected)

	input = `{
		a = ~b; // ok
	}
	`
	expected = `{
    a = ~b; // ok
}
`
	_testFormat(t, input, expected)

}

func TestFormatNewLines(t *testing.T) {
	input := "int foo() {\n    return 0;\n}\n\nint bar {\n    return 1;\n}\n"
	expected := "int foo() {\n    return 0;\n}\n\nint bar {\n    return 1;\n}\n"
	_testFormat(t, input, expected)

	input = "int foo() {\n    return 0;\n}int bar {\n    return 1;\n}\n"
	expected = "int foo() {\n    return 0;\n}\n\nint bar {\n    return 1;\n}\n"
	_testFormat(t, input, expected)

	input = "int foo() {\n    return 0;\n}\n\n\n\nint bar {\n    return 1;\n}\n"
	expected = "int foo() {\n    return 0;\n}\n\nint bar {\n    return 1;\n}\n"
	_testFormat(t, input, expected)

	input = "int foo() {\n    return 0;\n}\n\n\nint bar {\n    return 1;\n}\n\n\n"
	expected = "int foo() {\n    return 0;\n}\n\nint bar {\n    return 1;\n}\n"
	_testFormat(t, input, expected)

	input = "int foo() {\n    return 0;\n}\n\n\nint bar {\n    return 1;\n\n}\n"
	expected = "int foo() {\n    return 0;\n}\n\nint bar {\n    return 1;\n}\n"
	_testFormat(t, input, expected)

	input = "int foo() {\n    int i = 3;\n\n    return i;\n}\n\n\nint bar {\n    return 1;\n}\n"
	expected = "int foo() {\n    int i = 3;\n\n    return i;\n}\n\nint bar {\n    return 1;\n}\n"
	_testFormat(t, input, expected)

	input = "int foo() {\n    int i = 3;\n    return i;\n}\n\n\nint bar {\n    return 1;\n}\n"
	expected = "int foo() {\n    int i = 3;\n    return i;\n}\n\nint bar {\n    return 1;\n}\n"
	_testFormat(t, input, expected)

	input = "int foo() {\n    int i = 3;return i;\n}\n\n\nint bar {\n    return 1;\n}\n"
	expected = "int foo() {\n    int i = 3;\n    return i;\n}\n\nint bar {\n    return 1;\n}\n"
	_testFormat(t, input, expected)
}

func TestFormatSingleLineComment(t *testing.T) {
	input := "int i = 3;//comment\n"
	expected := "int i = 3; // comment\n"
	_testFormat(t, input, expected)

	input = "int i = 3;\n//comment\n"
	expected = "int i = 3;\n\n// comment\n"
	_testFormat(t, input, expected)

	input = "void foo() {\n    int i = 3;//comment\n}\n"
	expected = "void foo() {\n    int i = 3; // comment\n}\n"
	_testFormat(t, input, expected)

	input = "void foo() {\n    int i = 3;\n    //comment\n}\n"
	expected = "void foo() {\n    int i = 3;\n    // comment\n}\n"
	_testFormat(t, input, expected)

	input = "Foo foo = {\"123\", //A comment\n123};\n"
	expected = "Foo foo = {\n    \"123\", // A comment\n    123\n};\n"
	_testFormat(t, input, expected)

	input = "//Shift left"
	expected = "// Shift left\n"
	_testFormat(t, input, expected)

	input = "//    Shift left"
	expected = "// Shift left\n"
	_testFormat(t, input, expected)

}

func TestFormatMultilineLineComment(t *testing.T) {
	input := "/*comment*/"
	expected := "/*\n   comment\n*/\n"
	_testFormat(t, input, expected)

	input = "/*\n\ncomment\n\n*/"
	expected = "/*\n   comment\n*/\n"
	_testFormat(t, input, expected)

	input = "/*\n\ncomment\n\ncomment\n*/"
	expected = "/*\n   comment\n\n   comment\n*/\n"
	_testFormat(t, input, expected)

}

func TestFormatMacro(t *testing.T) {
	input := "#define MACRO(num, str) {printf(\"%d\", num);printf(\" is\");printf(\" %s number\", str);printf(\"\\n\");}\n"
	expected := "#define MACRO(num, str) {\\\n    printf(\"%d\", num);\\\n    printf(\" is\");\\\n    printf(\" %s number\", str);\\\n    printf(\"\\n\");\\\n}\n"
	_testFormat(t, input, expected)

	input = `#define MACRO(num, str) {\
    printf("%d", num);\
    printf(" is");\
    printf(" %s number", str);\
    printf("\n");\
}
`
	expected = `#define MACRO(num, str) {\
    printf("%d", num);\
    printf(" is");\
    printf(" %s number", str);\
    printf("\n");\
}
`
	_testFormat(t, input, expected)

	input = "#define MACRO(str) {\\\n    printf(\"%s\", str);\\\n}\\\n"
	expected = "#define MACRO(str) {\\\n    printf(\"%s\", str);\\\n}\n"
	_testFormat(t, input, expected)

	input = `// stringizer.c
#include <stdio.h>
#define stringer( x ) printf_s( #x "\n" )
int main() {
   stringer( In quotes in the printf function call );
   stringer( "In quotes when printed to the screen" );
   stringer( "This: \"  prints an escaped double quote" );
}`

	expected = `// stringizer.c
#include <stdio.h>

#define stringer(x) printf_s(#x "\n")

int main() {
    stringer(In quotes in the printf function call);
    stringer("In quotes when printed to the screen");
    stringer("This: \"  prints an escaped double quote");
}
`
	_testFormat(t, input, expected)

	input = `#define F abc

#define B def

#define FB(arg) #arg

#define FB1(arg) FB(arg)
`

	expected = `#define F abc

#define B def

#define FB(arg) #arg

#define FB1(arg) FB(arg)
`
	_testFormat(t, input, expected)

	input = "#define makechar(x)  #@x"
	expected = "#define makechar(x) #@x\n"
	_testFormat(t, input, expected)
}

func TestFormatDirective(t *testing.T) {
	input := "#endif\nint i = 1;"
	expected := "#endif\n\nint i = 1;\n"
	_testFormat(t, input, expected)

	input = "#include <stdio.h>\n"
	expected = "#include <stdio.h>\n"
	_testFormat(t, input, expected)

	input = "{#ifdef FOO\n foo(); #else\nbar(); #endif\n}"
	expected = "{\n#ifdef FOO\n    foo();\n#else\n    bar();\n#endif\n}\n"
	_testFormat(t, input, expected)

	input = `typedef struct stbtt__active_edge {
    #if STBTT_RASTERIZER_VERSION==1
    int x, int direction;
    #elif STBTT_RASTERIZER_VERSION==2
float fx,fdx,fdy;
    #else
#error "Unrecognized value of STBTT_RASTERIZER_VERSION"
    #endif
} stbtt__active_edge;
`
	expected = `typedef struct stbtt__active_edge {
#if STBTT_RASTERIZER_VERSION == 1
    int x, int direction;
#elif STBTT_RASTERIZER_VERSION == 2
    float fx, fdx, fdy;
#else
#error "Unrecognized value of STBTT_RASTERIZER_VERSION"
#endif
} stbtt__active_edge;
`
	_testFormat(t, input, expected)

	input = "#define STBTT__OVER_MASK (STBTT_MAX_OVERSAMPLE - 1)"
	expected = "#define STBTT__OVER_MASK (STBTT_MAX_OVERSAMPLE - 1)\n"
	_testFormat(t, input, expected)

	input = "  #define STBTT_ifloor(x)    ((int) floor(x))\n"
	expected = "#define STBTT_ifloor(x) ((int) floor(x))\n"
	_testFormat(t, input, expected)

	input = `#define paster(n) printf_s("token" #n " = %d", token##n)`
	expected = `#define paster(n) printf_s("token" #n " = %d", token##n)
`
	_testFormat(t, input, expected)

}

func TestFormatBrackets(t *testing.T) {
	input := "foo[1]=2;"
	expected := "foo[1] = 2;\n"
	_testFormat(t, input, expected)

	input = "foo [ 1 ]=2;"
	expected = "foo[1] = 2;\n"
	_testFormat(t, input, expected)

}

func TestFormatPointerTypes(t *testing.T) {
	input := "C8_Keypad * keypad = &(global_state.keypad);\n"
	expected := "C8_Keypad *keypad = &(global_state.keypad);\n"
	_testFormat(t, input, expected)

	input = "void c8_load_from_file_dialog(C8_State*state) {\n    printf(\"Hi\");\n}\n"
	expected = "void c8_load_from_file_dialog(C8_State *state) {\n    printf(\"Hi\");\n}\n"
	_testFormat(t, input, expected)
}

func TestFunctionArguments(t *testing.T) {
	input := "void c8_glyph(C8_State *state, C8_Glyph glyph, float x, float y, float width, float height, C8_Rgba rgb)"
	expected := `void c8_glyph(
    C8_State *state,
    C8_Glyph glyph,
    float x,
    float y,
    float width,
    float height,
    C8_Rgba rgb
)
`
	_testFormat(t, input, expected)

	input = "void c8_glyph(C8_State *state, C8_Glyph glyph, float x, float y, float width, float height, C8_Rgba rgb){}"
	expected = `void c8_glyph(
    C8_State *state,
    C8_Glyph glyph,
    float x,
    float y,
    float width,
    float height,
    C8_Rgba rgb
) {
}
`
	_testFormat(t, input, expected)

	input = "bool c8_read_entire_file(const char *path, C8_Arena *arena, C8_File *read_result) {}"
	expected = "bool c8_read_entire_file(const char *path, C8_Arena *arena, C8_File *read_result) {\n}\n"
	_testFormat(t, input, expected)

	input = "bool foo(int a, int b, int c, int d, int e) {\n}\n"
	expected = "bool foo(int a, int b, int c, int d, int e) {\n}\n"
	_testFormat(t, input, expected)

	input = "bool foo(int a, int b, int c, int d, int e, int f) {\n}\n"
	expected = "bool foo(int a, int b, int c, int d, int e, int f) {\n}\n"
	_testFormat(t, input, expected)

	input = "int foo() {\n    return 0;\n}\n"
	expected = "int foo() {\n    return 0;\n}\n"
	_testFormat(t, input, expected)

	input = "void foo() {\n    bar();\n}\n"
	expected = "void foo() {\n    bar();\n}\n"
	_testFormat(t, input, expected)

}

func TestFormatFunctionCall(t *testing.T) {
	input := "{bool c8_read_entire_file(const char *path, C8_Arena *arena, C8_File *read_result);}\n"
	expected := `{
    bool c8_read_entire_file(
        const char *path,
        C8_Arena *arena,
        C8_File *read_result
    );
}
`
	_testFormat(t, input, expected)

	input = "{\n    foo();\n}\n"
	expected = "{\n    foo();\n}\n"
	_testFormat(t, input, expected)

	input = "{\n    foo(bar);\n}\n"
	expected = "{\n    foo(bar);\n}\n"
	_testFormat(t, input, expected)

	input = "{bool c8_read_entire_file(const char *path, C8_Arena *arena);}\n"
	expected = "{\n    bool c8_read_entire_file(const char *path, C8_Arena *arena);\n}\n"
	_testFormat(t, input, expected)

	input = "{c8_color_vertex(p1 .x, p1 .y);}\n"
	expected = "{\n    c8_color_vertex(p1.x, p1.y);\n}\n"
	_testFormat(t, input, expected)

	input = "{int c = foo(foo(1, 2), foo(3, 4), foo(5, 6), foo(5, 6), foo(5, 6), foo(5, 6), foo(5, 6) );}"
	expected = `{
    int c = foo(
        foo(1, 2),
        foo(3, 4),
        foo(5, 6),
        foo(5, 6),
        foo(5, 6),
        foo(5, 6),
        foo(5, 6)
    );
}
`
	_testFormat(t, input, expected)
}

func TestFormatWrapping(t *testing.T) {

	input := `{C8_Button load_button = state->load_button;
 state->load_button.is_mouse_over = state->mouse_position.x >= load_button.x
&& state->mouse_position.x <= load_button.x + load_button.width
&& state->mouse_position.y >= load_button.y
&& state->mouse_position.y <= load_button.y + load_button.height;}`
	expected := `{
    C8_Button load_button = state->load_button;
    state->load_button.is_mouse_over = state->mouse_position.x >= load_button.x
        && state->mouse_position.x <= load_button.x + load_button.width
        && state->mouse_position.y >= load_button.y
        && state->mouse_position.y <= load_button.y + load_button.height;
}
`

	_testFormat(t, input, expected)

	input = "{state->load_button.is_mouse_over = state->mouse_position.x >= load_button.x && state->mouse_position.x <= load_button.x + load_button.width && state->mouse_position.y >= load_button.y && state->mouse_position.y <= load_button.y + load_button.height;}"
	expected = `{
    state->load_button.is_mouse_over = state->mouse_position.x >= load_button.x && state->mouse_position.x <= load_button.x + load_button.width && state->mouse_position.y >= load_button.y && state->mouse_position.y <= load_button.y + load_button.height;
}
`
	_testFormat(t, input, expected)

}
