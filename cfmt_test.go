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
			t.Errorf("Index %d, expected %d, found end of string", i, r)
			break
		}

		if r != output[i] {
			t.Errorf("Index %d, expected %d (%x), output %d (%x)", i, r, r, output[i], output[i])
		}
	}

	if output != expected {
		t.Errorf("Output should be:\n%s\nfound:\n%s\n", expected, output)
	}
}

func TestTokenizeString(t *testing.T) {
	_testTokenizeSingleToken(t, "\"toto\"", String)
	_testTokenizeSingleToken(t, "\"to\\\"o\"", String)
}

func TestTokenizeFloat(t *testing.T) {
	_testTokenizeSingleToken(t, "55.0f", Float)
	_testTokenizeSingleToken(t, "123.456e-67", Float)
	_testTokenizeSingleToken(t, "123e+86", Float)
}

func TestFormatStructDecl(t *testing.T) {
	input :=
		"typedef struct {\r\n" +
			"    int bar;     char *baz;}Foo;"

	expected :=
		"typedef struct {\r\n" +
			"    int bar;\r\n" +
			"    char *baz;\r\n" +
			"} Foo;\r\n"

	_testFormat(t, input, expected)

	input =
		"struct  Foo{\r\n" +
			"    int bar;     char *baz;\r\n\r\n\r\n" +
			"}\r\n\r\n\r\n" +
			";"

	expected =
		"struct Foo {\r\n" +
			"    int bar;\r\n" +
			"    char *baz;\r\n" +
			"};\r\n"

	_testFormat(t, input, expected)

	input = "struct Foo{\r\n\r\n    int bar;\r\n\r\n    char *baz;\r\n\r\n};\r\n"

	expected = "struct Foo {\r\n    int bar;\r\n    char *baz;\r\n};\r\n"

	input = "typedef struct { struct {C8_Key kp_0;} keypad;} C8_Keypad;"
	expected = "typedef struct {\r\n    struct {\r\n        C8_Key kp_0;\r\n    } keypad;\r\n} C8_Keypad;\r\n"
	_testFormat(t, input, expected)

}

func TestFormatInitializerList(t *testing.T) {
	input := "Foo foo = {\r\n" +
		"0    }\r\n" +
		";"

	expected := "Foo foo = {0};\r\n"

	_testFormat(t, input, expected)

	input = " p = {.x,.y};\r\n"

	expected = "p = {.x, .y};\r\n"

	_testFormat(t, input, expected)

	input = " p = {. x, . y};\r\n"

	expected = "p = {.x, .y};\r\n"

	_testFormat(t, input, expected)

	input = " p = { .x, .y};\r\n"

	expected = "p = {.x, .y};\r\n"

	_testFormat(t, input, expected)

}

func TestFormatForLoop(t *testing.T) {
	input := " for (int i=0;i<3;i++) {printf(\"%d\\n\", i);}"

	expected := "for (int i = 0; i < 3; i++) {\r\n" +
		"    printf(\"%d\\n\", i);\r\n" +
		"}\r\n"

	_testFormat(t, input, expected)

	input = " for (int i=0;\r\ni<3;\r\ni++\r\n) {printf(\"%d\\n\", i);}"

	expected = "for (int i = 0; i < 3; i++) {\r\n" +
		"    printf(\"%d\\n\", i);\r\n" +
		"}\r\n"

	_testFormat(t, input, expected)

	input = "Foo zz = {123, \"123\"  };"
	expected = "Foo zz = {123, \"123\"};\r\n"
	_testFormat(t, input, expected)

	input = "Foo zz = {\r\n123,\r\n\"123\"\r\n};"
	expected = "Foo zz = {123, \"123\"};\r\n"
	_testFormat(t, input, expected)

}

func TestFormatOperators(t *testing.T) {
	input := "int a=b*c;"
	expected := "int a = b * c;\r\n"
	_testFormat(t, input, expected)

	input = "aa   -> bar =3;"
	expected = "aa->bar = 3;\r\n"
	_testFormat(t, input, expected)

	input = "a . b = c . d;"
	expected = "a.b = c.d;\r\n"
	_testFormat(t, input, expected)

	input = "a\r\n.b = c\r\n.d;"
	expected = "a.b = c.d;\r\n"
	_testFormat(t, input, expected)

	input = "i = i ++ == i --;\r\n"
	expected = "i = i++ == i--;\r\n"
	_testFormat(t, input, expected)

	input = "i = i\r\n++ == i\r\n--;\r\n"
	expected = "i = i++ == i--;\r\n"
	_testFormat(t, input, expected)

	input = "i = i++==i--;\r\n"
	expected = "i = i++ == i--;\r\n"
	_testFormat(t, input, expected)

	input = "i = i -- == i ++;\r\n"
	expected = "i = i-- == i++;\r\n"
	_testFormat(t, input, expected)

	input = "i = i\r\n-- == i\r\n++;\r\n"
	expected = "i = i-- == i++;\r\n"
	_testFormat(t, input, expected)

	input = "i = i--==i++;\r\n"
	expected = "i = i-- == i++;\r\n"
	_testFormat(t, input, expected)

	input = "i = (i) ++ == (i) --;\r\n"
	expected = "i = (i)++ == (i)--;\r\n"
	_testFormat(t, input, expected)

	input = "i = (i)\r\n++ == (i)\r\n--;\r\n"
	expected = "i = (i)++ == (i)--;\r\n"
	_testFormat(t, input, expected)

	input = "i = (i)++== (i)--;\r\n"
	expected = "i = (i)++ == (i)--;\r\n"
	_testFormat(t, input, expected)

	input = "i = (i)++ == (i)-- ;\r\n"
	expected = "i = (i)++ == (i)--;\r\n"
	_testFormat(t, input, expected)

	input = "i = ((i) ++)-- == ((i) ++)--;\r\n"
	expected = "i = ((i)++)-- == ((i)++)--;\r\n"
	_testFormat(t, input, expected)

	input = "i = ((i) --)-- == ((i) --) --;\r\n"
	expected = "i = ((i)--)-- == ((i)--)--;\r\n"
	_testFormat(t, input, expected)

	input = "i = (i)++<= (i)--;\r\n"
	expected = "i = (i)++ <= (i)--;\r\n"
	_testFormat(t, input, expected)

	input = "i = (i)++ <= (i)-- ;\r\n"
	expected = "i = (i)++ <= (i)--;\r\n"
	_testFormat(t, input, expected)

	input = "i = ((i) ++)-- <= ((i) ++)--;\r\n"
	expected = "i = ((i)++)-- <= ((i)++)--;\r\n"
	_testFormat(t, input, expected)

	input = "i = ((i) --)-- <= ((i) --) --;\r\n"
	expected = "i = ((i)--)-- <= ((i)--)--;\r\n"
	_testFormat(t, input, expected)

	input = "i = (i)++!= (i)--;\r\n"
	expected = "i = (i)++ != (i)--;\r\n"
	_testFormat(t, input, expected)

	input = "i = (i)++ != (i)-- ;\r\n"
	expected = "i = (i)++ != (i)--;\r\n"
	_testFormat(t, input, expected)

	input = "i = ((i) ++)-- != ((i) ++)--;\r\n"
	expected = "i = ((i)++)-- != ((i)++)--;\r\n"
	_testFormat(t, input, expected)

	input = "i = ((i) --)-- != ((i) --) --;\r\n"
	expected = "i = ((i)--)-- != ((i)--)--;\r\n"
	_testFormat(t, input, expected)
}

func TestFormatNewLines(t *testing.T) {
	input := "int foo() {\r\n    return 0;\r\n}\r\n\r\n\r\nint bar {\r\n    return 1;\r\n}\r\n"
	expected := "int foo() {\r\n    return 0;\r\n}\r\n\r\n\r\nint bar {\r\n    return 1;\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "int foo() {\r\n    return 0;\r\n}int bar {\r\n    return 1;\r\n}\r\n"
	expected = "int foo() {\r\n    return 0;\r\n}\r\n\r\n\r\nint bar {\r\n    return 1;\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "int foo() {\r\n    return 0;\r\n}\r\n\r\n\r\n\r\nint bar {\r\n    return 1;\r\n}\r\n"
	expected = "int foo() {\r\n    return 0;\r\n}\r\n\r\n\r\nint bar {\r\n    return 1;\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "int foo() {\r\n    return 0;\r\n}\r\n\r\n\r\nint bar {\r\n    return 1;\r\n}\r\n\r\n\r\n"
	expected = "int foo() {\r\n    return 0;\r\n}\r\n\r\n\r\nint bar {\r\n    return 1;\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "int foo() {\r\n    return 0;\r\n}\r\n\r\n\r\nint bar {\r\n    return 1;\r\n\r\n}\r\n"
	expected = "int foo() {\r\n    return 0;\r\n}\r\n\r\n\r\nint bar {\r\n    return 1;\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "int foo() {\r\n    int i = 3;\r\n\r\n    return i;\r\n}\r\n\r\n\r\nint bar {\r\n    return 1;\r\n}\r\n"
	expected = "int foo() {\r\n    int i = 3;\r\n\r\n    return i;\r\n}\r\n\r\n\r\nint bar {\r\n    return 1;\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "int foo() {\r\n    int i = 3;\r\n    return i;\r\n}\r\n\r\n\r\nint bar {\r\n    return 1;\r\n}\r\n"
	expected = "int foo() {\r\n    int i = 3;\r\n    return i;\r\n}\r\n\r\n\r\nint bar {\r\n    return 1;\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "int foo() {\r\n    int i = 3;return i;\r\n}\r\n\r\n\r\nint bar {\r\n    return 1;\r\n}\r\n"
	expected = "int foo() {\r\n    int i = 3;\r\n    return i;\r\n}\r\n\r\n\r\nint bar {\r\n    return 1;\r\n}\r\n"
	_testFormat(t, input, expected)
}

func TestFormatSingleLineComment(t *testing.T) {
	input := "int i = 3;//comment\r\n"
	expected := "int i = 3; //comment\r\n"
	_testFormat(t, input, expected)

	input = "int i = 3;\r\n//comment\r\n"
	expected = "int i = 3;\r\n\r\n\r\n//comment\r\n"
	_testFormat(t, input, expected)

	input = "void foo() {\r\n    int i = 3;//comment\r\n}\r\n"
	expected = "void foo() {\r\n    int i = 3; //comment\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "void foo() {\r\n    int i = 3;\r\n    //comment\r\n}\r\n"
	expected = "void foo() {\r\n    int i = 3;\r\n    //comment\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "Foo foo = {\"123\", //A comment\r\n123};\r\n"
	expected = "Foo foo = {\"123\", //A comment\r\n123};\r\n"
	_testFormat(t, input, expected)
}

func TestFormatMultilineLineComment(t *testing.T) {
	input := "/*comment*/"
	expected := "/*\r\n   comment\r\n*/\r\n"
	_testFormat(t, input, expected)

	input = "/*\r\n\r\ncomment\r\n\r\n*/"
	expected = "/*\r\n   comment\r\n*/\r\n"
	_testFormat(t, input, expected)

	input = "/*\r\n\r\ncomment\r\n\r\ncomment\r\n*/"
	expected = "/*\r\n   comment\r\n\r\n   comment\r\n*/\r\n"
	_testFormat(t, input, expected)

}

func TestFormatMacro(t *testing.T) {
	input := "#define MACRO(num, str) {printf(\"%d\", num);printf(\" is\");printf(\" %s number\", str);printf(\"\\n\");}\r\n"
	expected := "#define MACRO(num, str) {\\\r\n    printf(\"%d\", num);\\\r\n    printf(\" is\");\\\r\n    printf(\" %s number\", str);\\\r\n    printf(\"\\n\");\\\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "#define MACRO(num, str) {\\\r\n    printf(\"%d\", num);\\\r\n    printf(\" is\");\\\r\n    printf(\" %s number\", str);\\\r\n    printf(\"\\n\");\\\r\n}\r\n"
	expected = "#define MACRO(num, str) {\\\r\n    printf(\"%d\", num);\\\r\n    printf(\" is\");\\\r\n    printf(\" %s number\", str);\\\r\n    printf(\"\\n\");\\\r\n}\r\n"
	_testFormat(t, input, expected)
}

func TestFormatDirective(t *testing.T) {
	input := "#endif\r\nint i = 1;"
	expected := "#endif\r\n\r\n\r\nint i = 1;\r\n"
	_testFormat(t, input, expected)

	input = "#include <stdio.h>\r\n"
	expected = "#include <stdio.h>\r\n"
	_testFormat(t, input, expected)

	input = "{#ifdef FOO\r\n foo(); #else\r\nbar(); #endif\r\n}"
	expected = "{\r\n    #ifdef FOO\r\n    foo();\r\n    #else\r\n    bar();\r\n    #endif\r\n}\r\n"
	_testFormat(t, input, expected)
}
