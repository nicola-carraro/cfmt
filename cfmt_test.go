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

	input = " p = { x, {y,\r\n z}};\r\n"
	expected = "p = {\r\n    x, {\r\n        y,\r\n        z\r\n    }\r\n};\r\n"

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
	expected = "Foo zz = {\r\n    123,\r\n    \"123\"\r\n};\r\n"
	_testFormat(t, input, expected)

	input = "int foo(){for (i16 i = 0; i < buf_size / sizeof(i16); i++) {return bar;}}"
	expected = "int foo() {\r\n    for (i16 i = 0; i < buf_size / sizeof(i16); i++) {\r\n        return bar;\r\n    }\r\n}\r\n"
	_testFormat(t, input, expected)

}

func TestFormatOperators(t *testing.T) {
	input := "int a=b*c;"
	expected := "int a = b * c;\r\n"
	_testFormat(t, input, expected)

	input = "aa   -> bar =3;"
	expected = "aa->bar = 3;\r\n"
	_testFormat(t, input, expected)

	input = "OPENFILENAME file_name = {.hwndOwner = state.window};\r\n"
	expected = "OPENFILENAME file_name = {.hwndOwner = state.window};\r\n"
	_testFormat(t, input, expected)

	input = "	OPENFILENAME file_name = {.lStructSize = sizeof(file_name),.hwndOwner = state.window,.hInstance = state.instance,.lpstrFilter = \"Chip 8 rom (*.ch8)\\0*.ch8\\0All files (*.*)'\\0*.*\",.lpstrFile = path,.nMaxFile = C8_ARRCOUNT(path),.lpstrInitialDir = \"roms\"};"
	expected = "OPENFILENAME file_name = {\r\n    .lStructSize = sizeof(file_name),\r\n    .hwndOwner = state.window,\r\n    .hInstance = state.instance,\r\n    .lpstrFilter = \"Chip 8 rom (*.ch8)\\0*.ch8\\0All files (*.*)'\\0*.*\",\r\n    .lpstrFile = path,\r\n    .nMaxFile = C8_ARRCOUNT(path),\r\n    .lpstrInitialDir = \"roms\"\r\n};\r\n"

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
	expected := "int i = 3; // comment\r\n"
	_testFormat(t, input, expected)

	input = "int i = 3;\r\n//comment\r\n"
	expected = "int i = 3;\r\n\r\n\r\n// comment\r\n"
	_testFormat(t, input, expected)

	input = "void foo() {\r\n    int i = 3;//comment\r\n}\r\n"
	expected = "void foo() {\r\n    int i = 3; // comment\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "void foo() {\r\n    int i = 3;\r\n    //comment\r\n}\r\n"
	expected = "void foo() {\r\n    int i = 3;\r\n    // comment\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "Foo foo = {\"123\", //A comment\r\n123};\r\n"
	expected = "Foo foo = {\r\n    \"123\", // A comment\r\n    123\r\n};\r\n"
	_testFormat(t, input, expected)

	input = "//Shift left"
	expected = "// Shift left\r\n"
	_testFormat(t, input, expected)

	input = "//    Shift left"
	expected = "// Shift left\r\n"
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

	input = "#define MACRO(str) {\\\r\n    printf(\"%s\", str);\\\r\n}\\\r\n"
	expected = "#define MACRO(str) {\\\r\n    printf(\"%s\", str);\\\r\n}\r\n"
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
	expected = "{\r\n#ifdef FOO\r\n    foo();\r\n#else\r\n    bar();\r\n#endif\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "typedef struct stbtt__active_edge {\r\n    #if STBTT_RASTERIZER_VERSION==1\r\n    int x, int direction;\r\n    #elif STBTT_RASTERIZER_VERSION==2\r\nfloat fx,fdx,fdy;\r\n    #else\r\n#error \"Unrecognized value of STBTT_RASTERIZER_VERSION\"\r\n    #endif\r\n} stbtt__active_edge;\r\n"
	expected = "typedef struct stbtt__active_edge {\r\n#if STBTT_RASTERIZER_VERSION == 1\r\n    int x, int direction;\r\n#elif STBTT_RASTERIZER_VERSION == 2\r\n    float fx, fdx, fdy;\r\n#else\r\n#error \"Unrecognized value of STBTT_RASTERIZER_VERSION\"\r\n#endif\r\n} stbtt__active_edge;\r\n"
	_testFormat(t, input, expected)

	input = "#define STBTT__OVER_MASK (STBTT_MAX_OVERSAMPLE - 1)"
	expected = "#define STBTT__OVER_MASK (STBTT_MAX_OVERSAMPLE - 1)\r\n"
	_testFormat(t, input, expected)

	input = "  #define STBTT_ifloor(x)    ((int) floor(x))\r\n"
	expected = "#define STBTT_ifloor(x) ((int) floor(x))\r\n"
	_testFormat(t, input, expected)
}

func TestFormatBrackets(t *testing.T) {
	input := "foo[1]=2;"
	expected := "foo[1] = 2;\r\n"
	_testFormat(t, input, expected)

	input = "foo [ 1 ]=2;"
	expected = "foo[1] = 2;\r\n"
	_testFormat(t, input, expected)

}

func TestFormatPointerTypes(t *testing.T) {
	input := "C8_Keypad * keypad = &(global_state.keypad);\r\n"
	expected := "C8_Keypad *keypad = &(global_state.keypad);\r\n"
	_testFormat(t, input, expected)

	input = "void c8_load_from_file_dialog(C8_State*state) {\r\n    printf(\"Hi\");\r\n}\r\n"
	expected = "void c8_load_from_file_dialog(C8_State *state) {\r\n    printf(\"Hi\");\r\n}\r\n"
	_testFormat(t, input, expected)
}

func TestFunctionArguments(t *testing.T) {
	input := "void c8_glyph(C8_State *state, C8_Glyph glyph, float x, float y, float width, float height, C8_Rgba rgb)"
	expected := "void c8_glyph(\r\n    C8_State *state,\r\n    C8_Glyph glyph,\r\n    float x,\r\n    float y,\r\n    float width,\r\n    float height,\r\n    C8_Rgba rgb\r\n)\r\n"
	_testFormat(t, input, expected)

	input = "void c8_glyph(C8_State *state, C8_Glyph glyph, float x, float y, float width, float height, C8_Rgba rgb){}"
	expected = "void c8_glyph(\r\n    C8_State *state,\r\n    C8_Glyph glyph,\r\n    float x,\r\n    float y,\r\n    float width,\r\n    float height,\r\n    C8_Rgba rgb\r\n) {\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "bool c8_read_entire_file(const char *path, C8_Arena *arena, C8_File *read_result) {}"
	expected = "bool c8_read_entire_file(\r\n    const char *path,\r\n    C8_Arena *arena,\r\n    C8_File *read_result\r\n) {\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "bool foo(int a, int b, int c, int d, int e) {\r\n}\r\n"
	expected = "bool foo(\r\n    int a,\r\n    int b,\r\n    int c,\r\n    int d,\r\n    int e\r\n) {\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "bool foo(int a, int b, int c, int d, int e, int f) {\r\n}\r\n"
	expected = "bool foo(\r\n    int a,\r\n    int b,\r\n    int c,\r\n    int d,\r\n    int e,\r\n    int f\r\n) {\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "int foo() {\r\n    return 0;\r\n}\r\n"
	expected = "int foo() {\r\n    return 0;\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "void foo() {\r\n    bar();\r\n}\r\n"
	expected = "void foo() {\r\n    bar();\r\n}\r\n"
	_testFormat(t, input, expected)

}

func TestWrapStatement(t *testing.T) {
	input := "{state->load_button.is_mouse_over = state->mouse_position.x >= load_button.x && state->mouse_position.x <= load_button.x + load_button.width && state->mouse_position.y >= load_button.y && state->mouse_position.y <= load_button.y + load_button.height;}"
	expected := "{\r\n    state->load_button.is_mouse_over = state->mouse_position.x >= load_button.x && state->mouse_position.x <= load_button.x + load_button.width && state->mouse_position.y >= load_button.y && state->mouse_position.y <= load_button.y + load_button.height;\r\n}\r\n"
	_testFormat(t, input, expected)
}

func TestFormatFunctionCall(t *testing.T) {
	input := "{bool c8_read_entire_file(const char *path, C8_Arena *arena, C8_File *read_result);}\r\n"
	expected := "{\r\n    bool c8_read_entire_file(\r\n        const char *path,\r\n        C8_Arena *arena,\r\n        C8_File *read_result\r\n    );\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "{\r\n    foo();\r\n}\r\n"
	expected = "{\r\n    foo();\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "{\r\n    foo(bar);\r\n}\r\n"
	expected = "{\r\n    foo(bar);\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "{bool c8_read_entire_file(const char *path, C8_Arena *arena);}\r\n"
	expected = "{\r\n    bool c8_read_entire_file(const char *path, C8_Arena *arena);\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "{c8_color_vertex(p1 .x, p1 .y);}\r\n"
	expected = "{\r\n    c8_color_vertex(p1.x, p1.y);\r\n}\r\n"
	_testFormat(t, input, expected)

	input = "{int c = foo(foo(1, 2), foo(3, 4), foo(5, 6), foo(5, 6), foo(5, 6), foo(5, 6), foo(5, 6) );}"
	expected = "{\r\n    int c = foo(\r\n        foo(1, 2),\r\n        foo(3, 4),\r\n        foo(5, 6),\r\n        foo(5, 6),\r\n        foo(5, 6),\r\n        foo(5, 6),\r\n        foo(5, 6)\r\n    );\r\n}\r\n"
	_testFormat(t, input, expected)
}

func TestFormatWrapping(t *testing.T){

		input := "{C8_Button load_button = state->load_button;\r\n state->load_button.is_mouse_over = state->mouse_position.x >= load_button.x\r\n&& state->mouse_position.x <= load_button.x + load_button.width\r\n&& state->mouse_position.y >= load_button.y\r\n&& state->mouse_position.y <= load_button.y + load_button.height;}"
		expected := "{\r\n    C8_Button load_button = state->load_button;\r\n    state->load_button.is_mouse_over = state->mouse_position.x >= load_button.x\r\n        && state->mouse_position.x <= load_button.x + load_button.width\r\n        && state->mouse_position.y >= load_button.y\r\n        && state->mouse_position.y <= load_button.y + load_button.height;\r\n}\r\n"

		_testFormat(t, input, expected)

	}
