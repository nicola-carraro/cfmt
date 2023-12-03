#ifndef CHIP8_H


#define CHIP8_H


#undef UNICODE


#include < stdio.h >


#include < assert.h >


#include < stdint.h >


#include < stdbool.h >


#include < stdlib.h >


#include < Windows.h >


#include < d3d9.h >


#include < DSound.h >


#include "shobjidl.h"


#define C8_DEBUG_PRINT 1


#define C8_MAX_VERTICES(C8_PIXEL_ROWS * C8_PIXEL_COLS * 2 * 3) + (8 * 3)


typedef int 16 _t i16;


typedef int 32 _t i32;


typedef int 64 _t i64;


typedef uint32_t u32;


typedef uint64_t u64;


typedef uint16_t u16;


typedef size_t psz;


typedef uint8_t u8;


#define C8_UNREFERENCED(a) (a)


#define C8_DWORD_MAX 4294967295


#define C8_RAM_BYTE_SIZE 4096


#define C8_FRAME_WIDTH 4.0f


#define C8_PIXEL_SIDE 8.0f


#define C8_PIXEL_ROWS 32


#define C8_PIXEL_COLS 64


#define C8_SCREEN_WIDTH(C8_PIXEL_SIDE * C8_PIXEL_COLS)


#define C8_SCREEN_HEIGHT(C8_PIXEL_SIDE * C8_PIXEL_ROWS)


#define C8_MONITOR_WIDTH(C8_SCREEN_WIDTH + (2 * C8_FRAME_WIDTH))


#define C8_MONITOR_HEIGHT(C8_SCREEN_HEIGHT + (2 * C8_FRAME_WIDTH))


#define C8_ARRCOUNT(a) (sizeof (a) / sizeof (a [ 0 ]))


#define C8_INSTRUCTIONS_PER_SEC 880


#define C8_FRAMES_PER_SEC 60


#define C8_INSTRUCTIONS_PER_FRAME(C8_INSTRUCTIONS_PER_SEC / C8_FRAMES_PER_SEC)


#define C8_FONT_SIZE 5


#define C8_FONT_COUNT 16


#define C8_FONT_ADDR 0 x050


#define C8_PROG_ADDR 0 x200


#define C8_FLAG_REG 0 xf


#define C8_FIRST_CHAR 'A'


#define C8_ONE_PAST_LAST_CHAR 'z'


#define C8_CHAR_COUNT(C8_ONE_PAST_LAST_CHAR - C8_FIRST_CHAR)


#define C8_LOAD_BUTTON_WIDTH(150.0f)


#define C8_LOAD_BUTTON_HEIGHT(70.0f)


#define C8_LOAD_BUTTON_BORDER(3.0f)


#define C8_LOAD_BUTTON_Y(10.0f)


#define C8_LOAD_BUTTON_TITLE("Load")


#define C8_TEXT_HEIGHT(60.0f)


#define C8_TEXT_SPACING(3.0f)


#define C8_D3D_FVF(D3DFVF_XYZRHW | D3DFVF_DIFFUSE)


#define C8_TEX_FVF(D3DFVF_XYZRHW | D3DFVF_DIFFUSE | D3DFVF_TEX1)


#ifdef C8_TEST


#define C8_LOG_ERROR(msg) {\
}


#else


#define C8_LOG_ERROR(msg) c8_log_error(__FILE__, __LINE__, msg)


#endif


int i = 0;


typedef struct {
    u8 r;
    u8 g;
    u8 b;
    u8 a;
} C8_Rgba;


typedef struct {
    float x;
    float y;
    C8_Rgba color;
    float u;
    float v;
} C8_Texture_Vertex;


typedef struct {
    float x;
    float y;
    C8_Rgba color;
} C8_Color_Vertex;


typedef struct {
    float u_left;
    float u_right;
    float v_top;
    float v_bottom;
    float y_offset;
    float width;
    float height;
    float advancement;
    float ascent;
    float descent;
} C8_Glyph;


typedef struct {
    C8_Glyph glyphs [ C8_CHAR_COUNT ];
    uint32_t width;
    uint32_t height;
    float above_bl;
    float text_height;
} C8_Font;


typedef struct {
    void *data;
    psz max_bytes;
    i32 alignement;
    psz offset;
} C8_Arena;


typedef struct {
    bool started_down;
    bool ended_down;
    bool was_down;
    bool was_pressed;
    bool was_lifted;
    i32 half_transitions;
} C8_Key;


typedef union {
    struct {
        C8_Key kp_0;
        C8_Key kp_1;
        C8_Key kp_2;
        C8_Key kp_3;
        C8_Key kp_4;
        C8_Key kp_5;
        C8_Key kp_6;
        C8_Key kp_7;
        C8_Key kp_8;
        C8_Key kp_9;
        C8_Key kp_a;
        C8_Key kp_b;
        C8_Key kp_c;
        C8_Key kp_d;
        C8_Key kp_e;
        C8_Key kp_f;
    } keypad;
    C8_Key keys [ 16 ];
} C8_Keypad;


typedef union {
    struct {
        C8_Key enter;
        C8_Key esc;
        C8_Key p;
        C8_Key space;
    } control_keys;
    C8_Key keys [ 4 ];
} C8_Control_Keys;


typedef struct {
    C8_Key left_button;
    C8_Key right_button;
} C8_Mouse_Buttons;


typedef struct {
    float x;
    float y;
} C8_V2;


typedef struct {
    psz size;
    void *data;
} C8_File;


typedef struct {
    FLOAT x;
    FLOAT y;
    FLOAT z;
    FLOAT rhw;
    D3DCOLOR color;
    float u;
    float v;
} C8_D3D_Texture_Vertex;


typedef struct {
    FLOAT x;
    FLOAT y;
    FLOAT z;
    FLOAT rhw;
    D3DCOLOR color;
} C8_D3D_Color_Vertex;


typedef struct {
    LARGE_INTEGER perf_freq;
    LARGE_INTEGER perf_count;
    bool has_timer;
} C8_Timer;


typedef struct {
    float x;
    float y;
    float width;
    float height;
    float text_y_offset;
    float text_x_offset;
    float text_spacing;
    float text_scale;
    float border;
    C8_Rgba text_color;
    char *title;
    bool is_down;
    bool is_mouse_over;
} C8_Button;


typedef struct {
    bool running;
    i32 cli_width;
    i32 cli_height;
    bool pixels [ C8_PIXEL_ROWS ] [ C8_PIXEL_COLS ];
    u8 ram [ C8_RAM_BYTE_SIZE ];
    u8 delay_timer;
    u8 sound_timer;
    u16 index_register;
    u8 var_registers [ 16 ];
    u16 stack [ 16 ];
    u8 stack_pointer;
    u64 frame_count;
    C8_Arena arena;
    u16 pc;
    bool initialised;
    bool program_loaded;
    C8_Keypad keypad;
    C8_Control_Keys control_keys;
    bool should_beep;
    C8_Font font;
    C8_V2 mouse_position;
    C8_Mouse_Buttons mouse_buttons;
    C8_Color_Vertex color_vertices [ C8_MAX_VERTICES ];
    uint32_t color_vertex_count;
    C8_Texture_Vertex text_vertices [ C8_MAX_VERTICES ];
    uint32_t text_vertex_count;
    LPDIRECT3D9 d3d;
    LPDIRECT3DDEVICE9 d3d_dev;
    LPDIRECT3DVERTEXBUFFER9 color_vb;
    LPDIRECT3DVERTEXBUFFER9 text_vb;
    LPDIRECT3DTEXTURE9 texture;
    LPDIRECTSOUND ds;
    LPDIRECTSOUNDBUFFER ds_sec_buf;
    HWND window;
    HINSTANCE instance;
    C8_Button load_button;
    bool has_sound;
    bool is_beeping;
} C8_State;


static C8_State global_state;


void *c8_allocate(psz size);


void *c8_arena_alloc(C8_Arena * arena, psz size);


bool c8_arena_init(C8_Arena * arena, psz size, i32 alignement);


void c8_arena_free_all(C8_Arena * arena);


void c8_app_update(C8_State * state);


const C8_Rgba emulator_color = {0, 0, 0, 255};


void c8_color_vertex(C8_State * state, float x, float y, u8 r, u8 g, u8 b, u8 a);


void c8_rect(C8_State * state, float x, float y, float width, float height, C8_Rgba rgb);


void c8_glyph(C8_State * state, C8_Glyph glyph, float x, float y, float width, float height, C8_Rgba rgb);


void c8_text_vertex(C8_State * state, float x, float y, u8 r, u8 g, u8 b, u8 a, float u, float v);


bool c8_read_entire_file(const char *path, C8_Arena * arena, C8_File * read_result);


void c8_message_box(const char *message);


void c8_load_rom(const char *filePath, C8_State * state);


void c8_log_error(const char *file, uint32_t line, char const *const msg);


void c8_load_button_init(C8_State * state, C8_Button * button);


void c8_bad_rom(C8_State * state);


#endif // !CHIP8_H\

