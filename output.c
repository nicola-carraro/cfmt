#ifndef CHIP8_C


#define CHIP8_C


#define _CRT_SECURE_NO_WARNINGS


#define COBJMACROS


#include "chip8.h"


#pragma comment(lib, "kernel32.lib")


#pragma comment(lib, "user32.lib")


#pragma comment(lib, "D3d9.lib")


#pragma comment(lib, "Dsound.lib")


#pragma comment(lib, "Dsound.lib")


#pragma comment(lib, "ole32.lib")


#pragma comment(lib, "Comdlg32.lib")


void c8_render_text(C8_State * state) {
    HRESULT hr = IDirect3DDevice9_SetFVF(state->d3d_dev, C8_TEX_FVF);

    if (FAILED(hr)) {
        C8_LOG_ERROR("Could not set FVF for texture\n");
        assert(false);
    }

    hr = IDirect3DDevice9_SetStreamSource(state->d3d_dev, 0, state->text_vb, 0, sizeof (C8_D3D_Texture_Vertex));

    if (FAILED(hr)) {
        C8_LOG_ERROR("Could not set stream source texture\n");
        assert(false);
    }

    C8_D3D_Texture_Vertex * vertices;

    hr = IDirect3DVertexBuffer9_Lock(state->text_vb, 0, 0, (void **) & vertices, 0);
    if (SUCCEEDED(hr)) {
        for (size_t i = 0; i < state->text_vertex_count; i++) {
            C8_Texture_Vertex source_vertex = state->text_vertices [ i ];
            C8_D3D_Texture_Vertex vertex = {0};
            vertex.x = source_vertex.x;
            vertex.y = source_vertex.y;
            vertex.rhw = 1.0f;
            vertex.u = source_vertex.u;
            vertex.v = source_vertex.v;
            C8_Rgba source_color = source_vertex.color;
            vertex.color = D3DCOLOR_ARGB(source_color.a, source_color.r, source_color.g, source_color.b);
            vertices [ i ] = vertex;
        }

        hr = IDirect3DVertexBuffer9_Unlock(state->text_vb);
        if (FAILED(hr)) {
            C8_LOG_ERROR("Could not unlock vertex buffer\n");
            assert(false);
        }
    }
    else {
        C8_LOG_ERROR("Could not lock vertex buffer\n");
        assert(false);
    }

    hr = IDirect3DDevice9_SetRenderState(state->d3d_dev, D3DRS_LIGHTING, FALSE);
    if (FAILED(hr)) {
        C8_LOG_ERROR("Could not set D3DRS_LIGHTING\n");
        assert(false);
    }

    hr = IDirect3DDevice9_SetRenderState(state->d3d_dev, D3DRS_ALPHABLENDENABLE, TRUE);
    if (FAILED(hr)) {
        C8_LOG_ERROR("Could not set D3DRS_ALPHABLENDENABLE\n");
        assert(false);
    }

    hr = IDirect3DDevice9_SetRenderState(state->d3d_dev, D3DRS_SRCBLEND, D3DBLEND_SRCALPHA);
    if (FAILED(hr)) {
        C8_LOG_ERROR("Could not set D3DBLEND_SRCALPHA\n");
        assert(false);
    }

    hr = IDirect3DDevice9_SetRenderState(state->d3d_dev, D3DRS_DESTBLEND, D3DBLEND_INVSRCALPHA);
    if (FAILED(hr)) {
        C8_LOG_ERROR("Could not set D3DBLEND_INVSRCALPHA\n");
        assert(false);
    }

    hr = IDirect3DDevice9_SetRenderState(state->d3d_dev, D3DTSS_ALPHAOP, D3DTOP_MODULATE);
    if (FAILED(hr)) {
        C8_LOG_ERROR("Could not set D3DTOP_MODULATE\n");
        assert(false);
    }

    hr = IDirect3DDevice9_SetTexture(state->d3d_dev, 0, (IDirect3DBaseTexture9 *) state->texture);
    if (FAILED(hr)) {
        C8_LOG_ERROR("Could not set texture\n");
        assert(false);
    }

    hr = IDirect3DDevice9_DrawPrimitive(state->d3d_dev, D3DPT_TRIANGLELIST, 0, state->text_vertex_count);
    if (FAILED(hr)) {
        C8_LOG_ERROR("Could not draw primitive\n");

        assert(false);
    }

    state->text_vertex_count = 0;
}


BOOL c8_load_font(C8_State * state, const char *file_name) {
    BOOL result = false;

    C8_File file = {0};

    if (c8_read_entire_file(file_name, &state->arena, &file)) {
        C8_Font font = *((C8_Font *) file.data);

        state->font = font;

        u8 * input_bitmap = (u8 *) file.data + sizeof (font);
        HRESULT textureCreated = IDirect3DDevice9_CreateTexture(state->d3d_dev, font.width, font.height, 0, 0, D3DFMT_A32B32G32R32F, D3DPOOL_MANAGED, &state->texture, NULL);

        if (SUCCEEDED(textureCreated)) {
            D3DLOCKED_RECT out_rect = {0};

            HRESULT locked = IDirect3DTexture9_LockRect(state->texture, 0, &out_rect, 0, 0);
            if (SUCCEEDED(locked)) {
                uint8_t * row_start = (uint8_t *) (out_rect.pBits);
                for (uint32_t row = 0; row < font.height; row++) {
                    for (uint32_t column = 0; column < font.width; column++) {
                        uint8_t src_pixel = input_bitmap [ (row * font.width) + column ];

                        float alpha = ((float) src_pixel) / 255.0f;

                        ((float *) row_start) [ column * 4 ] = 1.0f;
                        ((float *) row_start) [ column * 4 + 1 ] = 1.0f;
                        ((float *) row_start) [ column * 4 + 2 ] = 1.0f;
                        ((float *) row_start) [ column * 4 + 3 ] = alpha;
                    }

                    row_start += out_rect.Pitch;
                }
            }
            else {
                C8_LOG_ERROR("Could not lock rectangle\n");
            }

            input_bitmap = ((uint8_t *) file.data + 2);

            HRESULT unlocked = IDirect3DTexture9_UnlockRect(state->texture, 0);
            if (SUCCEEDED(unlocked)) {
                HRESULT vb_created = IDirect3DDevice9_CreateVertexBuffer(state->d3d_dev, sizeof (state->text_vertices), 0, C8_TEX_FVF, D3DPOOL_MANAGED, &state->text_vb, 0);

                if (SUCCEEDED(vb_created)) {
                    result = TRUE;
                }
                else {
                    C8_LOG_ERROR("Could not create vertex buffer\n");
                }
            }
            else {
                C8_LOG_ERROR("Could not unlock rect\n");
            }
        }
        else {
            C8_LOG_ERROR("Could not create texture\n");
        }
    }
    else {
        C8_LOG_ERROR("Could not load font\n");
    }

    return result;
}


D3DPRESENT_PARAMETERS c8_init_d3d_params(HWND window) {
    D3DPRESENT_PARAMETERS result = {0};

    result.Windowed = TRUE;
    result.SwapEffect = D3DSWAPEFFECT_DISCARD;

    result.BackBufferWidth = 0;
    result.BackBufferHeight = 0;
    result.BackBufferFormat = D3DFMT_UNKNOWN;
    result.BackBufferCount = 0;
    result.MultiSampleType = D3DMULTISAMPLE_NONE;
    result.MultiSampleQuality = 0;
    result.SwapEffect = D3DSWAPEFFECT_DISCARD;
    result.hDeviceWindow = window;
    result.Windowed = true;
    result.EnableAutoDepthStencil = 0;
    result.AutoDepthStencilFormat = D3DFMT_UNKNOWN;
    result.Flags = 0;
    result.FullScreen_RefreshRateInHz = 0;
    result.PresentationInterval = D3DPRESENT_INTERVAL_ONE;

    return result;
}


bool c8_query_perf_count(C8_Timer * timer, LARGE_INTEGER * perf_count) {
    bool result = false;
    if (QueryPerformanceCounter(perf_count)) {
        result = true;
    }
    else {
        timer->has_timer = false;
        C8_LOG_ERROR("Could not get performance count\n");
    }
    return result;
}


C8_Timer c8_init_timer() {
    C8_Timer result = {0};

    LARGE_INTEGER perf_freq = {0};

    if (QueryPerformanceFrequency(&perf_freq)) {
        LARGE_INTEGER perf_count;

        if (c8_query_perf_count(&result, &perf_count)) {
            result.has_timer = true;
            result.perf_freq = perf_freq;
            result.perf_count = perf_count;
        }
    }
    else {
        C8_LOG_ERROR("Could not get performance frequency\n");
    }

    return result;
}


float c8_compute_millis(C8_Timer timer, LARGE_INTEGER new_perf_count) {
    float secs_elapsed = ((float) (new_perf_count.QuadPart - timer.perf_count.QuadPart)) / ((float) (timer.perf_freq.QuadPart));

    float result = secs_elapsed * 1000.0f;

    return result;
}


float c8_millis_elapsed(C8_Timer * timer, bool reset_timer) {
    float millis_elapsed = - 1.0f;

    if (timer->has_timer) {
        LARGE_INTEGER perf_count;

        if (c8_query_perf_count(timer, &perf_count)) {
            millis_elapsed = c8_compute_millis(*timer, perf_count);

            if (reset_timer) {
                timer->perf_count = perf_count;
            }
        }
    }

    return millis_elapsed;
}


void c8_render_color(C8_State * state) {
    VOID * vp;

    HRESULT hr = IDirect3DDevice9_SetTexture(state->d3d_dev, 0, 0);
    if (FAILED(hr)) {
        C8_LOG_ERROR("Could not unset texture\n");
        assert(false);
    }

    hr = IDirect3DVertexBuffer9_Lock(state->color_vb, 0, 0, &vp, 0);

    if (SUCCEEDED(hr)) {
        for (size_t vertex_index = 0; vertex_index < state->color_vertex_count; vertex_index++) {
            C8_D3D_Color_Vertex win_vertex = {0};
            C8_Color_Vertex vertex = state->color_vertices [ vertex_index ];
            win_vertex.x = vertex.x;
            win_vertex.y = vertex.y;
            win_vertex.rhw = 1.0f;
            C8_Rgba color = vertex.color;
            win_vertex.color = D3DCOLOR_ARGB(color.a, color.r, color.g, color.b);
            ((C8_D3D_Color_Vertex *) vp) [ vertex_index ] = win_vertex;
        }

        hr = IDirect3DVertexBuffer9_Unlock(state->color_vb);
        if (FAILED(hr)) {
            C8_LOG_ERROR("Failed to unlock vertex buffer\n");

            assert(false);
        }
    }
    else {
        C8_LOG_ERROR("Failed to lock vertex buffer\n");
        assert(false);
    }

    hr = IDirect3DDevice9_SetFVF(state->d3d_dev, C8_D3D_FVF);
    if (FAILED(hr)) {
        C8_LOG_ERROR("SetFVF failed\n");
        assert(false);
    }

    hr = IDirect3DDevice9_SetStreamSource(state->d3d_dev, 0, state->color_vb, 0, sizeof (C8_D3D_Color_Vertex));
    if (FAILED(hr)) {
        C8_LOG_ERROR("SetStreamSource failed\n");
        assert(false);
    }

    hr = IDirect3DDevice9_DrawPrimitive(state->d3d_dev, D3DPT_TRIANGLELIST, 0, state->color_vertex_count / 3);

    if (FAILED(hr)) {
        C8_LOG_ERROR("DrawPrimitive failed\n");
        assert(false);
    }
}


void c8_render(C8_State * state) {
    HRESULT hr = IDirect3DDevice9_Clear(state->d3d_dev, 0, 0, D3DCLEAR_TARGET, D3DCOLOR_ARGB(255, 255, 255, 255), 1.0f, 0);

    if (FAILED(hr)) {
        C8_LOG_ERROR("Could not clear screen\n");
        assert(false);
    }

    hr = IDirect3DDevice9_BeginScene(state->d3d_dev);
    if (FAILED(hr)) {
        C8_LOG_ERROR("BeginScene failed\n");
        assert(false);
    }

    c8_render_color(state);

    c8_render_text(state);

    hr = IDirect3DDevice9_EndScene(state->d3d_dev);
    if (FAILED(hr)) {
        C8_LOG_ERROR("EndScene failed\n");
        assert(false);
    }

    hr = IDirect3DDevice9_Present(state->d3d_dev, 0, 0, 0, 0);

    if (FAILED(hr)) {
        C8_LOG_ERROR("Present failed\n");
        assert(false);
    }
}


bool c8_init_texture(C8_State * state, const char *file_name) {
    bool result = false;

    C8_File file = {0};

    if (c8_read_entire_file(file_name, &state->arena, &file)) {
        UINT width = ((UINT *) file.data) [ 0 ];
        UINT height = ((UINT *) file.data) [ 1 ];

        HRESULT textureCreated = IDirect3DDevice9_CreateTexture(state->d3d_dev, width, height, 0, 0, D3DFMT_A32B32G32R32F, D3DPOOL_MANAGED, &state->texture, NULL);

        if (SUCCEEDED(textureCreated)) {
            D3DLOCKED_RECT rect;
            rect.Pitch = width;
            rect.pBits = 0;

            u8 * input_bitmap = (u8 *) (((UINT *) file.data) + 2);
            HRESULT locked = IDirect3DTexture9_LockRect(state->texture, 0, &rect, 0, 0);

            if (SUCCEEDED(locked)) {
                u32 float _count = 0;
                for (u32 row = 0; row < height; row++) {
                    for (u32 column = 0; column < width; column++) {
                        u8 pixel = input_bitmap [ (row * width) + column ];

                        for (u32 channel = 0; channel < 4; channel++) {
                            float color = (float) pixel / 255.0f;
                            ((float *) rect.pBits) [ float _count ] = color;
                            float _count++;
                        }
                    }
                }

                HRESULT unlocked = IDirect3DTexture9_UnlockRect(state->texture, 0);

                if (SUCCEEDED(unlocked)) {
                    result = true;
                }
                else {
                    C8_LOG_ERROR("Failed to unlock texture\n");
                }
            }
            else {
                C8_LOG_ERROR("Failed to lock texture\n");
            }
        }
        else {
            C8_LOG_ERROR("Failed to create texture\n");
        }
    }

    else {
        C8_LOG_ERROR("Could not read texture file\n");
    }

    return result;
}


bool c8_initd3d(C8_State * state, HWND window) {
    state->d3d = Direct3DCreate9(DIRECT3D_VERSION);
    if (! state->d3d) {
        C8_LOG_ERROR("Could not create D3D interface\n");
        return false;
    }

    D3DPRESENT_PARAMETERS d3dpp = c8_init_d3d_params(window);
    HRESULT hr = IDirect3D9_CreateDevice(state->d3d, D3DADAPTER_DEFAULT, D3DDEVTYPE_HAL, window, D3DCREATE_HARDWARE_VERTEXPROCESSING, &d3dpp, &state->d3d_dev);

    if (FAILED(hr)) {
        C8_LOG_ERROR("Could not create D3D device\n");
        return false;
    }

    hr = IDirect3DDevice9_CreateVertexBuffer(state->d3d_dev, sizeof (state->color_vertices), 0, C8_D3D_FVF, D3DPOOL_MANAGED, &state->color_vb, 0);

    if (FAILED(hr)) {
        C8_LOG_ERROR("Could not create vertex buffer\n");
        return false;
    }

    const char file_name [ ] = "fonts/font";

    if (! c8_load_font(state, file_name)) {
        C8_LOG_ERROR("Could not load font\n");
        return false;
    }
    hr = IDirect3DDevice9_SetRenderState(state->d3d_dev, D3DRS_LIGHTING, FALSE);

    if (FAILED(hr)) {
        C8_LOG_ERROR("Could notset render state\n");
        return false;
    }

    hr = IDirect3DDevice9_SetRenderState(state->d3d_dev, D3DRS_ALPHABLENDENABLE, TRUE);
    if (FAILED(hr)) {
        C8_LOG_ERROR("Could notset render state\n");
        return false;
    }
    hr = IDirect3DDevice9_SetRenderState(state->d3d_dev, D3DRS_SRCBLEND, D3DBLEND_SRCALPHA);

    if (FAILED(hr)) {
        C8_LOG_ERROR("Could notset render state\n");
        return false;
    }
    hr = IDirect3DDevice9_SetRenderState(state->d3d_dev, D3DRS_DESTBLEND, D3DBLEND_INVSRCALPHA);

    if (FAILED(hr)) {
        C8_LOG_ERROR("Could notset render state\n");
        return false;
    }
    hr = IDirect3DDevice9_SetRenderState(state->d3d_dev, D3DTSS_ALPHAOP, D3DTOP_MODULATE);
    if (FAILED(hr)) {
        C8_LOG_ERROR("Could notset render state\n");
        return false;
    }

    return true;
}


void c8_process_key(C8_Key * key, WORD key_flags) {
    BOOL is_up = (key_flags & KF_UP) == KF_UP;
    if (is_up) {
        key->ended_down = false;
        key->was_lifted = true;
    }
    else {
        key->ended_down = true;
        key->was_down = true;
        key->was_pressed = true;
    }

    key->half_transitions++;
}


bool c8_init_dsound(C8_State * state, HWND window, i32 samples_per_sec) {
    bool result = false;

    LPDIRECTSOUND dsound;

    HRESULT ds_created = DirectSoundCreate(0, &dsound, 0);

    WAVEFORMATEX wformat;
    wformat.nChannels = 1;

    DWORD buf_bytes = samples_per_sec * sizeof (i16) * wformat.nChannels;

    wformat.wFormatTag = WAVE_FORMAT_PCM;
    wformat.nSamplesPerSec = samples_per_sec;
    wformat.wBitsPerSample = 16;
    wformat.nBlockAlign = (wformat.nChannels * wformat.wBitsPerSample) / 8;
    wformat.nAvgBytesPerSec = wformat.nSamplesPerSec * wformat.nBlockAlign;
    wformat.cbSize = 8;

    if (SUCCEEDED(ds_created)) {
        HRESULT cl = IDirectSound_SetCooperativeLevel(dsound, window, DSSCL_PRIORITY);
        if (SUCCEEDED(cl)) {
            DSBUFFERDESC pbuf_desc = {0};
            pbuf_desc.dwSize = sizeof (pbuf_desc);
            pbuf_desc.dwFlags = DSBCAPS_PRIMARYBUFFER;
            LPDIRECTSOUNDBUFFER pbuf;

            HRESULT pbuf_created = IDirectSound_CreateSoundBuffer(dsound, &pbuf_desc, &pbuf, 0);

            if (SUCCEEDED(pbuf_created)) {
                HRESULT fset = IDirectSoundBuffer_SetFormat(pbuf, &wformat);
                if (SUCCEEDED(fset)) {
                    DSBUFFERDESC sbuf_desc = {0};
                    sbuf_desc.dwSize = sizeof (sbuf_desc);
                    sbuf_desc.dwBufferBytes = buf_bytes;
                    sbuf_desc.lpwfxFormat = &wformat;
                    LPDIRECTSOUNDBUFFER sbuf;

                    HRESULT sbuf_created = IDirectSound_CreateSoundBuffer(dsound, &sbuf_desc, &sbuf, 0);

                    if (SUCCEEDED(sbuf_created)) {
                        LPVOID buf_data;
                        DWORD buf_size;
                        HRESULT locked = IDirectSoundBuffer_Lock(sbuf, 0, 0, &buf_data, &buf_size, 0, 0, DSBLOCK_ENTIREBUFFER);

                        if (SUCCEEDED(locked)) {
                            i32 hz = 250;
                            i32 wave_period = samples_per_sec / hz;

                            i32 half_wave_period = wave_period / 2;

                            i16 * samples = (i16 *) buf_data;

                            for (i16 i = 0; i < buf_size / sizeof (i16);
                            i++) {
                                i16 value;
                                if ((i / half_wave_period) % 2 == 0) {
                                    value = 16000;
                                }
                                else {
                                    value = - 16000;
                                }

                                samples [ i ] = value;
                            }

                            HRESULT unlocked = IDirectSoundBuffer_Unlock(sbuf, buf_data, buf_size, 0, 0);

                            if (SUCCEEDED(unlocked)) {
                                state->ds = dsound;
                                state->ds_sec_buf = sbuf;
                                state->has_sound = true;
                                result = true;
                            }
                            else {
                                C8_LOG_ERROR("Could not unlock buffer\n");
                            }
                        }
                        else {
                            C8_LOG_ERROR("Could not lock buffer\n");
                        }
                    }
                    else {
                        C8_LOG_ERROR("Could not create secondary buffer\n");
                    }
                }
                else {
                    C8_LOG_ERROR("Could not set wave format of buffer\n");
                }
            }
            else {
                C8_LOG_ERROR("Failed to create primary buffer\n");
            }
        }
        else {
            C8_LOG_ERROR("Failed to set cooperative level\n");
        }
    }
    else {
        C8_LOG_ERROR("Could not create dsound interface\n");
    }

    return result;
}


LRESULT CALLBACK WindowProc(HWND window, UINT msg, WPARAM wparam, LPARAM lparam) {
    switch (msg) {
        case WM_CLOSE : {
            if (! DestroyWindow(window)) {
                C8_LOG_ERROR("Could not destroy window\n");
            }
        }
        break;

        case WM_DESTROY : {
            PostQuitMessage(0);
            global_state.running = false;
        }
        break;
        case WM_QUIT : {
            global_state.running = false;
        }
        break;
        case WM_SIZE : {
            global_state.cli_width = LOWORD(lparam);
            global_state.cli_height = HIWORD(lparam);

            if (global_state.d3d_dev != 0) {
                D3DPRESENT_PARAMETERS d3dpp = c8_init_d3d_params(window);
                IDirect3DDevice9_Reset(global_state.d3d_dev, &d3dpp);
            }
        }
        break;
        case WM_LBUTTONDOWN : {
            global_state.mouse_buttons.left_button.ended_down = true;
            global_state.mouse_buttons.left_button.was_pressed = true;
        }
        break;
        case WM_LBUTTONUP : {
            global_state.mouse_buttons.left_button.ended_down = false;
            global_state.mouse_buttons.left_button.was_lifted = true;
        }
        break;
        case WM_RBUTTONDOWN : {
            global_state.mouse_buttons.right_button.ended_down = true;
            global_state.mouse_buttons.right_button.was_pressed = true;
        }
        break;
        case WM_RBUTTONUP : {
            global_state.mouse_buttons.right_button.ended_down = false;
            global_state.mouse_buttons.right_button.was_lifted = true;
        }
        break;
        case WM_KEYDOWN : case WM_KEYUP : {
            WORD key_flags = HIWORD(lparam);
            WORD scan_code = LOBYTE(key_flags);
            WORD vkey_code = LOWORD(wparam);
            C8_Control_Keys * controls = &(global_state.control_keys);
            switch (vkey_code) {
                case VK_RETURN : {
                    C8_Key * key = &(controls->control_keys.enter);
                    c8_process_key(key, key_flags);
                }
                break;
                case VK_ESCAPE : {
                    C8_Key * key = &(controls->control_keys.esc);
                    c8_process_key(key, key_flags);
                }
                break;
                case 'P' : {
                    C8_Key * key = &(controls->control_keys.p);
                    c8_process_key(key, key_flags);
                }
                break;
                case VK_SPACE : {
                    C8_Key * key = &(controls->control_keys.space);
                    c8_process_key(key, key_flags);
                }
                break;
            }

            C8_Keypad * keypad = &(global_state.keypad);
            switch (scan_code) {
                case 2 : {
                    C8_Key * key = &(keypad->keypad.kp_1);
                    c8_process_key(key, key_flags);
                }
                break;
                case 3 : {
                    C8_Key * key = &(keypad->keypad.kp_2);
                    c8_process_key(key, key_flags);
                }
                break;
                case 4 : {
                    C8_Key * key = &(keypad->keypad.kp_3);
                    c8_process_key(key, key_flags);
                }
                break;
                case 5 : {
                    C8_Key * key = &(keypad->keypad.kp_c);
                    c8_process_key(key, key_flags);
                }
                break;
                case 16 : {
                    C8_Key * key = &(keypad->keypad.kp_4);
                    c8_process_key(key, key_flags);
                }
                break;
                case 17 : {
                    C8_Key * key = &(keypad->keypad.kp_5);
                    c8_process_key(key, key_flags);
                }
                break;
                case 18 : {
                    C8_Key * key = &(keypad->keypad.kp_6);
                    c8_process_key(key, key_flags);
                }
                break;
                case 19 : {
                    C8_Key * key = &(keypad->keypad.kp_d);
                    c8_process_key(key, key_flags);
                }
                break;
                case 30 : {
                    C8_Key * key = &(keypad->keypad.kp_7);
                    c8_process_key(key, key_flags);
                }
                break;
                case 31 : {
                    C8_Key * key = &(keypad->keypad.kp_8);
                    c8_process_key(key, key_flags);
                }
                break;
                case 32 : {
                    C8_Key * key = &(keypad->keypad.kp_9);
                    c8_process_key(key, key_flags);
                }
                break;
                case 33 : {
                    C8_Key * key = &(keypad->keypad.kp_e);
                    c8_process_key(key, key_flags);
                }
                break;
                case 44 : {
                    C8_Key * key = &(keypad->keypad.kp_a);
                    c8_process_key(key, key_flags);
                }
                break;
                case 45 : {
                    C8_Key * key = &(keypad->keypad.kp_0);
                    c8_process_key(key, key_flags);
                }
                break;
                case 46 : {
                    C8_Key * key = &(keypad->keypad.kp_b);
                    c8_process_key(key, key_flags);
                }
                break;
                case 47 : {
                    C8_Key * key = &(keypad->keypad.kp_f);
                    c8_process_key(key, key_flags);
                }
                break;
            }
        }
        break;
    }

    return DefWindowProcA(window, msg, wparam, lparam);
}


void c8_log(const char *message) {
    printf(message);
    OutputDebugString(message);
}


void c8_logln(const char *message) {
    c8_log(message);
    c8_log("\n");
}


void c8_log_error(const char *file, uint32_t line, const char *msg) {
    char buf [ 1024 ] = {0};
    snprintf(buf, sizeof (buf), "ERROR (%s:%u): %s", file, line, msg);
    c8_log(buf);
}


HWND c8_create_window(HINSTANCE instance, int width, int height) {
    const char *class_name = "chip8";

    WNDCLASS wc = {.lpfnWndProc = WindowProc, .hInstance = instance, .lpszClassName = class_name};

    HWND window = 0;
    if (RegisterClass(&wc) != 0) {
        window = CreateWindowEx(0, class_name, "Chip 8", WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU, CW_USEDEFAULT, CW_USEDEFAULT, width, height, 0, 0, instance, 0);

        if (! window) {
            C8_LOG_ERROR("Could not create window\n");
        }
    }
    else {
        C8_LOG_ERROR("Could not register window class\n");
    }

    return window;
}


bool c8_process_msgs(C8_State * state, HWND window) {
    MSG msg;
    while (PeekMessage(&msg, window, 0, 0, PM_REMOVE)) {
        TranslateMessage(&msg);
        DispatchMessageA(&msg);
    }

    if (! state->running) {
        return false;
    }

    return true;
}


void c8_color_triangle(C8_State * state, C8_V2 p1, C8_V2 p2, C8_V2 p3, C8_Rgba rgb) {
    c8_color_vertex(state, p1.x, p1.y, rgb.r, rgb.g, rgb.b, rgb.a);
    c8_color_vertex(state, p2.x, p2.y, rgb.r, rgb.g, rgb.b, rgb.a);
    c8_color_vertex(state, p3.x, p3.y, rgb.r, rgb.g, rgb.b, rgb.a);
}


bool c8_start_beep(C8_State * state) {
    bool result = false;

    HRESULT play = IDirectSoundBuffer_Play(state->ds_sec_buf, 0, 0, DSBPLAY_LOOPING);
    if (SUCCEEDED(play)) {
        result = true;
        state->is_beeping = true;
    }
    else {
        C8_LOG_ERROR("Could not start beeping\n");
    }

    return result;
}


bool c8_stop_beep(C8_State * state) {
    bool result = false;

    HRESULT stop = IDirectSoundBuffer_Stop(state->ds_sec_buf);
    if (SUCCEEDED(stop)) {
        result = true;
        state->is_beeping = false;
    }
    else {
        C8_LOG_ERROR("Could not stop beeping\n");
    }

    return result;
}


wchar_t * c8_get_first_argument(LPWSTR cmd_line, C8_Arena * arena) {
    wchar_t * first_argument = 0;

    size_t length = 0;
    size_t start = 0;

    wchar_t * next_char = cmd_line;
    while (*next_char != '\0' && isspace(*next_char)) {
        next_char++;
        start++;
    }

    while (*next_char != '\0' && ! isspace(*next_char)) {
        next_char++;
        length++;
    }

    if (length > 0) {
        first_argument = c8_arena_alloc(arena, sizeof (*first_argument) * (length + 1));
        wcsncpy(first_argument, cmd_line + start, length);

        assert(first_argument && "Allocation failed");
    }

    return first_argument;
}


void c8_load_from_file_dialog(C8_State * state) {
    char path [ 1024 ] = {0};

    OPENFILENAME file_name = {.lStructSize = sizeof (file_name), .hwndOwner = state->window, .hInstance = state->instance, .lpstrFilter = "Chip 8 rom (*.ch8)\0*.ch8\0All files (*.*)'\0*.*", .lpstrFile = path, .nMaxFile = C8_ARRCOUNT(path), .lpstrInitialDir = "roms"};

    if (GetOpenFileName(&file_name)) {
        c8_load_rom(path, state);
    }
    else {
        DWORD error = CommDlgExtendedError();
        if (error) {
            c8_message_box("Error while opening file");
        }
    }
}


void c8_message_box(const char *message) {
    #ifndef C8_TEST BOOL succeded = MessageBox(global_state.window, message, "Error", MB_OK);

    if (! succeded) {
        C8_LOG_ERROR("Error while showing message box\n");
    }

    #else C8_UNREFERENCED(message);
    #endif }


int WINAPI WinMain(HINSTANCE instance, HINSTANCE prev_instance, PSTR cmd_line, int cmd_show) {
    C8_UNREFERENCED(cmd_line);

    C8_UNREFERENCED(prev_instance);

    HWND window = c8_create_window(instance, CW_USEDEFAULT, CW_USEDEFAULT);

    const char initError [ ] = "Fatal error while initialising application";

    if (! window) {
        c8_message_box(initError);
        C8_LOG_ERROR("Could not open window'\n");
        return - 1;
    }

    global_state.window = window;
    global_state.instance = instance;

    C8_Timer timer = c8_init_timer();

    i32 samples_per_sec = 8000;

    if (! c8_arena_init(&(global_state.arena), 5 * 1024 * 1024, 4)) {
        c8_message_box(initError);
        C8_LOG_ERROR("Could not initialise arena\n");
        return - 1;
    }

    if (! c8_initd3d(&global_state, window)) {
        c8_message_box(initError);
        C8_LOG_ERROR("Could not initialise Direct3D\n");
        return - 1;
    }

    ShowWindow(window, cmd_show);
    c8_load_button_init(&global_state, &global_state.load_button);

    global_state.running = true;

    global_state.has_sound = false;
    if (c8_init_dsound(&global_state, window, samples_per_sec)) {
        global_state.has_sound = true;
    }
    else {
        C8_LOG_ERROR("Could not initialize Direct Sound\n");
        assert(false);
    }

    while (global_state.running) {
        if (! c8_process_msgs(&global_state, window)) {
            break;
        }

        WINDOWPLACEMENT window_placement;
        window_placement.length = sizeof (WINDOWPLACEMENT);
        if (GetWindowPlacement(window, &window_placement)) {
            POINT point;
            if (GetCursorPos(&point) && ScreenToClient(window, &point)) {
                global_state.mouse_position.x = (float) point.x;
                global_state.mouse_position.y = (float) point.y;
            }
            else {
                C8_LOG_ERROR("Could not get mouse position\n");
                assert(false);
            }
        }
        else {
            C8_LOG_ERROR("Could not get client position\n");
            assert(false);
        }

        c8_app_update(&global_state);

        c8_render(&global_state);

        if (! global_state.is_beeping && global_state.should_beep) {
            bool beeped = c8_start_beep(&global_state);
            assert(beeped);
        }

        if (global_state.is_beeping && ! global_state.should_beep) {
            bool stopped = c8_stop_beep(&global_state);
            assert(stopped);
        }

        float milli_elapsed = c8_millis_elapsed(&timer, true);

        char str [ 255 ];
        sprintf(str, "Milliseconds: %f\n", milli_elapsed);
        // c8_log(str);
    }

    fflush(stdout);

    return 0;
}


void *c8_allocate(psz size) {
    void *result = VirtualAlloc(0, size, MEM_COMMIT, PAGE_READWRITE);

    return result;
}


HANDLE c8_open_file_for_read(const char *path) {
    HANDLE f = CreateFile(path, GENERIC_READ, FILE_SHARE_READ, 0, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, 0);

    if (f == INVALID_HANDLE_VALUE) {
        C8_LOG_ERROR("Could not open file");
        c8_logln(path);
    }

    return f;
}


uint64_t c8_file_size(HANDLE file) {
    uint64_t result = 0;
    LARGE_INTEGER f_size = {0};
    if (GetFileSizeEx(file, &f_size)) {
        result = f_size.QuadPart;
    }
    else {
        C8_LOG_ERROR("Could not get size of file");
    }

    return result;
}


bool c8_read_file(HANDLE f, char *buffer, uint64_t total_bytes_to_read) {
    while (total_bytes_to_read > 0) {
        DWORD bytes_to_read;
        if (total_bytes_to_read <= MAXDWORD) {
            bytes_to_read = (DWORD) total_bytes_to_read;
        }
        else {
            bytes_to_read = MAXDWORD;
        }

        DWORD bytes_read;

        if (! ReadFile(f, buffer, bytes_to_read, &bytes_read, 0)) {
            return false;
        }

        total_bytes_to_read -= bytes_read;
    }

    return true;
}


bool c8_close_file(HANDLE file) {
    return CloseHandle(file);
}


bool c8_read_entire_file(const char *path, C8_Arena * arena, C8_File * read_result) {
    bool result = false;
    char buf [ 256 ];

    HANDLE f = c8_open_file_for_read(path);
    if (f != INVALID_HANDLE_VALUE) {
        uint64_t size = c8_file_size(f);

        void *data = c8_arena_alloc(arena, size);

        if (data != 0) {
            if (c8_read_file(f, data, size)) {
                read_result->size = size;
                read_result->data = data;
                result = true;
            }
            else {
                C8_LOG_ERROR("Could not read file: ");
                c8_logln(path);
            }
        }
        else {
            C8_LOG_ERROR("Could not allocate memory for file");
        }

        if (! c8_close_file(f)) {
            snprintf(buf, sizeof (buf) - 1, "Could not close %s\n", path);
            C8_LOG_ERROR(buf);
        }
    }
    else {
        C8_LOG_ERROR("Could not open file: ");
        c8_logln(path);
    }

    return result;
}


u16 c8_read_instruction(u16 bytes) {
    u16 result = ((bytes & 0 x00ff) << 8) | ((bytes & 0 xff00) >> 8);
    return result;
}


float c8_frame_x(C8_State * state) {
    float result = (state->cli_width / 2) - (C8_MONITOR_WIDTH / 2);

    return result;
}


float c8_frame_y(C8_State * state) {
    float result = (state->cli_height / 2) - (C8_MONITOR_HEIGHT / 2);

    return result;
}


void c8_emulator_frame(C8_State * state) {
    float frame_x = c8_frame_x(state);

    float frame_y = c8_frame_y(state);

    c8_rect(state, frame_x, frame_y, C8_MONITOR_WIDTH, C8_FRAME_WIDTH, emulator_color);

    c8_rect(state, frame_x + C8_MONITOR_WIDTH, frame_y, C8_FRAME_WIDTH, C8_MONITOR_HEIGHT, emulator_color);

    c8_rect(state, frame_x + C8_FRAME_WIDTH, frame_y + C8_MONITOR_HEIGHT, C8_MONITOR_WIDTH, C8_FRAME_WIDTH, emulator_color);

    c8_rect(state, frame_x, frame_y + C8_FRAME_WIDTH, C8_FRAME_WIDTH, C8_MONITOR_HEIGHT, emulator_color);
}


void c8_color_vertex(C8_State * state, float x, float y, u8 r, u8 g, u8 b, u8 a) {
    assert(state->color_vertex_count < C8_ARRCOUNT(state->color_vertices));

    if (state->color_vertex_count < C8_ARRCOUNT(state->color_vertices)) {
        state->color_vertices [ state->color_vertex_count ].x = x;
        state->color_vertices [ state->color_vertex_count ].y = y;

        C8_Rgba color = {0};
        color.r = r;
        color.g = g;
        color.b = b;
        color.a = a;
        state->color_vertices [ state->color_vertex_count ].color = color;
        state->color_vertex_count++;
    }
    else {
        C8_LOG_ERROR("Vertex buffer size exceeded\n");
    }
}


void c8_reset_emulator(C8_State * state) {
    memset(state->ram, 0, sizeof (state->ram));
    memset(state->pixels, 0, sizeof (state->pixels));
    state->pc = C8_PROG_ADDR;
}


void c8_load_rom(const char *path, C8_State * state) {
    HANDLE handle = c8_open_file_for_read(path);

    if (handle == INVALID_HANDLE_VALUE) {
        C8_LOG_ERROR("Could not read rom: ");
        c8_logln(path);
        goto cleanup;
    }

    uint64_t size = c8_file_size(handle);

    if (size > sizeof (state->ram)) {
        c8_message_box("File is too large for emulator's RAM\n");
        goto cleanup;
    }

    char *buffer = c8_arena_alloc(&state->arena, size);

    if (! buffer) {
        C8_LOG_ERROR("Could not allocate memory for ROM\n");
        goto cleanup;
    }

    if (c8_read_file(handle, buffer, size)) {
        c8_reset_emulator(state);
        memcpy(state->ram + state->pc, buffer, size);

        state->program_loaded = true;

        c8_arena_free_all(&state->arena);

        const u8 font_sprites [ C8_FONT_SIZE * C8_FONT_COUNT ] = {0 xF0, 0 x90, 0 x90, 0 x90, 0 xF0, // 0
        0 x20, 0 x60, 0 x20, 0 x20, 0 x70, // 1
        0 xF0, 0 x10, 0 xF0, 0 x80, 0 xF0, // 2
        0 xF0, 0 x10, 0 xF0, 0 x10, 0 xF0, // 3
        0 x90, 0 x90, 0 xF0, 0 x10, 0 x10, // 4
        0 xF0, 0 x80, 0 xF0, 0 x10, 0 xF0, // 5
        0 xF0, 0 x80, 0 xF0, 0 x90, 0 xF0, // 6
        0 xF0, 0 x10, 0 x20, 0 x40, 0 x40, // 7
        0 xF0, 0 x90, 0 xF0, 0 x90, 0 xF0, // 8
        0 xF0, 0 x90, 0 xF0, 0 x10, 0 xF0, // 9
        0 xF0, 0 x90, 0 xF0, 0 x90, 0 x90, // A
        0 xE0, 0 x90, 0 xE0, 0 x90, 0 xE0, // B
        0 xF0, 0 x80, 0 x80, 0 x80, 0 xF0, // C
        0 xE0, 0 x90, 0 x90, 0 x90, 0 xE0, // D
        0 xF0, 0 x80, 0 xF0, 0 x80, 0 xF0, // E
        0 xF0, 0 x80, 0 xF0, 0 x80, 0 x80 // F
        };

        memcpy(state->ram + C8_FONT_ADDR, font_sprites, sizeof (font_sprites));
    }
    else {
        C8_LOG_ERROR("Could not read file: ");
        c8_logln(path);

        c8_message_box("Could not read rom");
    }

    cleanup : if (handle != INVALID_HANDLE_VALUE) {
        c8_close_file(handle);
    }
}


void c8_text_triangle(C8_State * state, C8_V2 p1, C8_V2 p2, C8_V2 p3, C8_Rgba rgb, float u1, float v1, float u2, float v2, float u3, float v3) {
    c8_text_vertex(state, p1.x, p1.y, rgb.r, rgb.g, rgb.b, rgb.a, u1, v1);
    c8_text_vertex(state, p2.x, p2.y, rgb.r, rgb.g, rgb.b, rgb.a, u2, v2);
    c8_text_vertex(state, p3.x, p3.y, rgb.r, rgb.g, rgb.b, rgb.a, u3, v3);
}


void c8_glyph(C8_State * state, C8_Glyph glyph, float x, float y, float width, float height, C8_Rgba rgb) {
    c8_text_vertex(state, x, y, rgb.r, rgb.g, rgb.b, rgb.a, glyph.u_left, glyph.v_top);
    c8_text_vertex(state, x + width, y, rgb.r, rgb.g, rgb.b, rgb.a, glyph.u_right, glyph.v_top);
    c8_text_vertex(state, x + width, y + height, rgb.r, rgb.g, rgb.b, rgb.a, glyph.u_right, glyph.v_bottom);

    c8_text_vertex(state, x, y, rgb.r, rgb.g, rgb.b, rgb.a, glyph.u_left, glyph.v_top);
    c8_text_vertex(state, x + width, y + height, rgb.r, rgb.g, rgb.b, rgb.a, glyph.u_right, glyph.v_bottom);
    c8_text_vertex(state, x, y + height, rgb.r, rgb.g, rgb.b, rgb.a, glyph.u_left, glyph.v_bottom);
}


void c8_text(C8_State * state, char *text, float x, float y, float scale, float spacing, C8_Rgba rgba) {
    float x_offset = 0;

    char c = 0;

    while (*text) {
        c = *text;
        C8_Glyph glyph = state->font.glyphs [ c - C8_FIRST_CHAR ];
        c8_glyph(state, glyph, x + x_offset, y + glyph.y_offset * scale, glyph.width * scale, glyph.height * scale, rgba);
        x_offset += glyph.width * scale;
        x_offset += spacing * scale;
        text++;
    }
}


void c8_text_vertex(C8_State * state, float x, float y, u8 r, u8 g, u8 b, u8 a, float u, float v) {
    assert(state->text_vertex_count < C8_ARRCOUNT(state->text_vertices));

    if (state->text_vertex_count < C8_ARRCOUNT(state->text_vertices)) {
        state->text_vertices [ state->text_vertex_count ].x = x;
        state->text_vertices [ state->text_vertex_count ].y = y;
        state->text_vertices [ state->text_vertex_count ].u = u;
        state->text_vertices [ state->text_vertex_count ].v = v;
        C8_Rgba color = {0};
        color.r = r;
        color.g = g;
        color.b = b;
        color.a = a;
        state->text_vertices [ state->text_vertex_count ].color = color;
        state->text_vertex_count++;
    }
    else {
        C8_LOG_ERROR("Vertex buffer size exceeded");
    }
}


void c8_rect(C8_State * state, float x, float y, float width, float height, C8_Rgba rgb) {
    C8_V2 p1 = {x, y};
    C8_V2 p2 = {x + width, y};
    C8_V2 p3 = {x + width, y + height};
    C8_V2 p4 = {x, y + height};

    c8_color_triangle(state, p1, p2, p3, rgb);
    c8_color_triangle(state, p1, p3, p4, rgb);
}


void c8_emulator_pixels(C8_State * state) {
    for (i32 r = 0; r < C8_ARRCOUNT(state->pixels);
    r++) {
        for (i32 c = 0; c < C8_ARRCOUNT(state->pixels [ r ]);
        c++) {
            if (state->pixels [ r ] [ c ]) {
                float frame_x = c8_frame_x(state);
                float frame_y = c8_frame_y(state);

                c8_rect(state, frame_x + (c * C8_PIXEL_SIDE) + C8_FRAME_WIDTH, frame_y + (r * C8_PIXEL_SIDE) + C8_FRAME_WIDTH, C8_PIXEL_SIDE, C8_PIXEL_SIDE, emulator_color);
            }
        }
    }
}


void c8_reset_key(C8_Key * k) {
    k->started_down = k->ended_down;
    k->was_down = k->ended_down;
    k->was_lifted = false;
    k->was_pressed = false;
    k->half_transitions = 0;
}


void c8_add_number_to_register(C8_State * state, u8 x, u8 nn) {
    u8 vx = state->var_registers [ x ];
    u16 result = (u16) vx + (u16) nn;

    if (result > 0 xff) {
        state->var_registers [ C8_FLAG_REG ] = 1;
    }
    else {
        state->var_registers [ C8_FLAG_REG ] = 0;
    }

    state->var_registers [ x ] = (result & 0 xff);
}


void c8_call(C8_State * state, u16 nnn) {
    state->stack [ state->stack_pointer ] = state->pc;
    state->stack_pointer++;
    if (state->stack_pointer >= C8_ARRCOUNT(state->stack)) {
        c8_bad_rom(state);
    }
    state->pc = nnn;
}


float c8_max_v_height(char *text, size_t text_length, C8_Font * font) {
    float result = 0.0f;
    for (size_t i = 0; i < text_length; i++) {
        char c = text [ i ];

        i32 glyph_index = c - C8_FIRST_CHAR;
        C8_Glyph glyph = font->glyphs [ glyph_index ];

        float v_height = glyph.v_bottom - glyph.v_top;

        if (v_height > result) {
            result = v_height;
        }
    }

    return result;
}


float c8_text_width(C8_Font * font, char *text, float text_scale, float spacing) {
    float result = 0.0f;

    float scaled_spacing = spacing * text_scale;
    while (*text) {
        char c = *text;
        C8_Glyph glyph = font->glyphs [ c - C8_FIRST_CHAR ];
        result += glyph.width * text_scale;
        result += scaled_spacing;
        text++;
    }

    if (result > 0.0f) {
        result -= scaled_spacing;
    }

    return result;
}


float c8_offset_to_center_vertically(C8_Font * font, const char *text, float text_scale, float container_height) {
    C8_UNREFERENCED(container_height);

    float max_ascent = 0.0f;
    float max_descent = 0.0f;
    while (*text) {
        char c = *text;
        C8_Glyph glyph = font->glyphs [ c - C8_FIRST_CHAR ];
        float scaled_ascent = glyph.ascent * text_scale;
        float scaled_descent = glyph.descent * text_scale;
        if (scaled_ascent > max_ascent) {
            max_ascent = scaled_ascent;
        }
        if (scaled_descent > max_descent) {
            max_descent = scaled_descent;
        }
        text++;
    }

    float text_effective_height = max_descent + max_ascent;

    float top_padding = (container_height - text_effective_height) / 2.0f;

    float result = top_padding + max_ascent;
    return result;
}


void c8_load_button_init(C8_State * state, C8_Button * button) {
    button->x = (state->cli_width / 2.0f) - (C8_LOAD_BUTTON_WIDTH / 2.0f);

    button->y = C8_LOAD_BUTTON_Y;
    button->width = C8_LOAD_BUTTON_WIDTH;
    button->height = C8_LOAD_BUTTON_HEIGHT;

    button->text_spacing = C8_TEXT_SPACING;

    button->text_color.a = 255;

    button->text_scale = C8_TEXT_HEIGHT / state->font.text_height;

    button->border = C8_LOAD_BUTTON_BORDER;

    button->title = C8_LOAD_BUTTON_TITLE;

    // float text_max_height = c8_text_max_height(&state->font, text, text_height);
    button->text_y_offset = c8_offset_to_center_vertically(&state->font, button->title, button->text_scale, button->height);

    float text_width = c8_text_width(&state->font, button->title, button->text_scale, button->text_spacing);
    button->text_x_offset = (button->width - text_width) / 2.0f;
}


void c8_load_button(C8_State * state) {
    C8_Button button = state->load_button;

    if (button.is_mouse_over) {
        c8_rect(state, button.x, button.y, button.width, button.height, button.text_color);

        C8_Rgba white = {255, 255, 255, 255};

        c8_text(state, button.title, button.x + button.text_x_offset, button.y + button.text_y_offset, button.text_scale, button.text_spacing, white);
    }
    else {
        c8_rect(state, button.x, button.y, button.width, button.border, button.text_color);

        c8_rect(state, button.x, button.y, button.border, button.height, button.text_color);

        c8_rect(state, button.x, button.y + button.height - button.border, button.width, button.border, button.text_color);

        c8_rect(state, button.x + button.width - button.border, button.y, button.border, button.height, button.text_color);

        c8_text(state, button.title, button.x + button.text_x_offset, button.y + button.text_y_offset, button.text_scale, button.text_spacing, button.text_color);
    }
}


void c8_bad_rom(C8_State * state) {
    c8_message_box("Bad ROM");
    c8_reset_emulator(state);
    state->program_loaded = false;
}


void c8_clear(C8_State * state) {
    memset(state->pixels, 0, sizeof (state->pixels));
}


void c8_return(C8_State * state) {
    state->stack_pointer--;

    if (state->stack_pointer > C8_ARRCOUNT(state->stack)) {
        C8_LOG_ERROR("Stack underflow\n");

        c8_bad_rom(state);
    }

    state->pc = state->stack [ state->stack_pointer ];
}


void c8_random(C8_State * state, u8 x, u8 nn) {
    int random = rand();
    u8 result = (u8) ((u16) random & nn);
    state->var_registers [ x ] = result;
}


void c8_display_sprite(C8_State * state, u8 x, u8 y, u8 n) {
    u8 flag_register = 0;
    u16 sprite_x = state->var_registers [ x ] % C8_PIXEL_COLS;
    u16 sprite_y = state->var_registers [ y ] % C8_PIXEL_ROWS;
    u8 * sprite_start = state->ram + state->index_register;

    for (i32 r = 0; r < n && sprite_y + r < C8_PIXEL_ROWS; r++) {
        u8 sprite_row = *(sprite_start + r);

        for (i32 c = 0; c < 8 && sprite_x + c < C8_PIXEL_COLS; c++) {
            u8 on = (sprite_row >> (7 - c)) & 0 x01;
            if (on == 0 x01) {
                if (state->pixels [ sprite_y + r ] [ sprite_x + c ]) {
                    state->pixels [ sprite_y + r ] [ sprite_x + c ] = false;
                }
                else {
                    state->pixels [ sprite_y + r ] [ sprite_x + c ] = true;
                }
                flag_register = 1;
            }
            else {
                assert(on == 0 x00);
            }
        }
    }

    state->var_registers [ C8_FLAG_REG ] = flag_register;
}


void c8_update_emulator(C8_State * state) {
    C8_Button load_button = state->load_button;

    state->load_button.is_mouse_over = state->mouse_position.x >= load_button.x && state->mouse_position.x <= load_button.x + load_button.width && state->mouse_position.y >= load_button.y && state->mouse_position.y <= load_button.y + load_button.height;

    if (state->mouse_buttons.left_button.was_pressed && state->load_button.is_mouse_over) {
        state->load_button.is_down = true;
    }

    if (state->load_button.is_down && state->mouse_buttons.left_button.was_lifted) {
        if (state->load_button.is_mouse_over) {
            c8_load_from_file_dialog(state);
        }
    }

    if (state->load_button.is_down && (state->mouse_buttons.left_button.was_lifted || ! (state->load_button.is_mouse_over))) {
        state->load_button.is_down = false;
    }

    c8_load_button(state);

    if (state->program_loaded) {
        for (i32 i = 0; i < C8_INSTRUCTIONS_PER_FRAME; i++) {
            // Break in case of bad ROM
            if (! state->program_loaded) {
                break;
            }

            assert(state->pc < sizeof (state->ram));
            u16 instruction = c8_read_instruction(*((u16 *) (state->ram + state->pc)));

            u8 op = instruction >> 12;

            u8 x = (instruction & 0 x0f00) >> 8;
            u8 y = (instruction & 0 x00f0) >> 4;
            u8 n = (instruction & 0 x000f);
            u8 nn = (instruction & 0 x00ff);
            u16 nnn = (instruction & 0 x0fff);

            u8 vx = state->var_registers [ x ];
            u8 vy = state->var_registers [ y ];

            state->pc += 2;

            if (instruction == 0 x00e0) {
                c8_clear(state);
            }
            else if (op == 0 x1) {
                // Jump
                state->pc = nnn;
            }
            else if (op == 0 x2) {
                c8_call(state, nnn);
            }
            else if (instruction == 0 x00ee) {
                c8_return(state);
            }
            else if (op == 0 x6) {
                // Set register
                state->var_registers [ x ] = nn;
            }
            else if (op == 0 x7) {
                // Add number to register
                c8_add_number_to_register(state, x, nn);
            }
            else if (op == 0 xa) {
                // Set index to register
                state->index_register = nnn;
            }
            else if (op == 0 xc) {
                c8_random(state, x, nn);
            }
            else if (op == 0 xd) {
                c8_display_sprite(state, x, y, n);
            }
            // Skip
            else if (op == 0 x3) {
                if (vx == nn) {
                    state->pc += 2;
                }
            }
            else if (op == 0 x4) {
                if (vx != nn) {
                    state->pc += 2;
                }
            }
            else if ((instruction & 0 xf00f) == 0 x5000) {
                if (state->var_registers [ x ] == state->var_registers [ y ]) {
                    state->pc += 2;
                }
            }
            else if ((instruction & 0 xf00f) == 0 x9000) {
                if (state->var_registers [ x ] != state->var_registers [ y ]) {
                    state->pc += 2;
                }
            }
            else if ((instruction & 0 xf00f) == 0 x8000) {
                // Set register to register
                state->var_registers [ x ] = state->var_registers [ y ];
            }
            else if ((instruction & 0 xf00f) == 0 x8001) {
                // Binary or
                state->var_registers [ x ] = state->var_registers [ x ] | state->var_registers [ y ];
            }
            else if ((instruction & 0 xf00f) == 0 x8002) {
                u8 result = vx & vy;
                // Binary and
                state->var_registers [ x ] = result;
            }
            else if ((instruction & 0 xf00f) == 0 x8003) {
                // Binary xor
                state->var_registers [ x ] = state->var_registers [ x ] ^ state->var_registers [ y ];
            }
            else if ((instruction & 0 xf00f) == 0 x8004) {
                // Add
                u16 result = state->var_registers [ x ] + state->var_registers [ y ];
                if (result > 0 xff) {
                    state->var_registers [ C8_FLAG_REG ] = 1;
                }
                else {
                    state->var_registers [ C8_FLAG_REG ] = 0;
                }
                state->var_registers [ x ] = (u8) result;
            }
            else if ((instruction & 0 xf00f) == 0 x8006) {
                // Shift right
                u8 bit = state->var_registers [ x ] &0 x1;
                state->var_registers [ x ] = state->var_registers [ x ] >> 1;
                state->var_registers [ C8_FLAG_REG ] = bit;
            }
            else if ((instruction & 0 xf00f) == 0 x8005) {
                // x - y
                if (state->var_registers [ x ] > state->var_registers [ y ]) {
                    state->var_registers [ C8_FLAG_REG ] = 1;
                }

                if (state->var_registers [ x ] < state->var_registers [ y ]) {
                    state->var_registers [ C8_FLAG_REG ] = 0;
                }
                state->var_registers [ x ] = state->var_registers [ x ] - state->var_registers [ y ];
            }
            else if ((instruction & 0 xf00f) == 0 x8007) {
                // y - x
                if (state->var_registers [ y ] > state->var_registers [ x ]) {
                    state->var_registers [ C8_FLAG_REG ] = 1;
                }

                if (state->var_registers [ y ] < state->var_registers [ x ]) {
                    state->var_registers [ C8_FLAG_REG ] = 0;
                }
                state->var_registers [ x ] = state->var_registers [ y ] - state->var_registers [ x ];
            }
            else if ((instruction & 0 xf00f) == 0 x800e) {
                // Shift left
                u8 bit = state->var_registers [ x ] >> 7;
                state->var_registers [ x ] = state->var_registers [ x ] << 1;
                state->var_registers [ C8_FLAG_REG ] = bit;
            }
            else if ((instruction & 0 xf0ff) == 0 xE09E) {
                // Skip if key pressed
                u8 key = state->var_registers [ x ];

                if (state->keypad.keys [ key ].ended_down) {
                    state->pc += 2;
                }
            }
            else if ((instruction & 0 xf0ff) == 0 xE0A1) {
                // Skip if key is not pressed
                u8 key = state->var_registers [ x ];
                if (! state->keypad.keys [ key ].ended_down) {
                    state->pc += 2;
                }
            }
            else if ((instruction & 0 xf0ff) == 0 xF00A) {
                // Wait for key
                if (! state->keypad.keys [ x ].ended_down) {
                    state->pc -= 2;
                }
            }
            else if ((instruction & 0 xf0ff) == 0 xF007) {
                // Get delay timer
                state->var_registers [ x ] = state->delay_timer;
            }
            else if ((instruction & 0 xf0ff) == 0 xF015) {
                // Set delay timer
                state->delay_timer = state->var_registers [ x ];
            }
            else if ((instruction & 0 xf0ff) == 0 xF018) {
                // Set sound timer
                state->sound_timer = state->var_registers [ x ];
            }
            else if ((instruction & 0 xf0ff) == 0 xF01E) {
                // Add to index
                u16 result = state->index_register + vx;

                if (result > 0 x0fff) {
                    state->var_registers [ C8_FLAG_REG ] = 1;
                    result &= 0 x0fff;
                }

                state->index_register = result;
            }
            else if ((instruction & 0 xf0ff) == 0 xF029) {
                // Point index to font
                u8 c = (state->var_registers [ x ]) & 0 x0f;
                state->index_register = C8_FONT_ADDR + (C8_FONT_SIZE * c);
            }
            else if ((instruction & 0 xf0ff) == 0 xF033) {
                // Decimal conversion
                u16 start = state->index_register;
                u8 dividend = state->var_registers [ x ];
                u8 divisor = 100;
                for (size_t digit_i = 0; digit_i < 3; digit_i++) {
                    state->ram [ start + digit_i ] = dividend / divisor;
                    dividend %= divisor;
                    divisor /= 10;
                }
            }
            else if ((instruction & 0 xf0ff) == 0 xF055) {
                // Store register
                u16 start = state->index_register;

                for (int reg_i = 0; reg_i <= x; reg_i++) {
                    state->ram [ start + reg_i ] = state->var_registers [ reg_i ];
                }
            }
            else if ((instruction & 0 xf0ff) == 0 xF065) {
                // Load register
                u16 start = state->index_register;

                for (i32 reg_i = 0; reg_i <= x; reg_i++) {
                    state->var_registers [ reg_i ] = state->ram [ start + reg_i ];
                }
            }
            else {
                c8_bad_rom(state);
                C8_LOG_ERROR("Unimplemented instruction\n");
            }
        }
    }

    c8_emulator_frame(state);

    c8_emulator_pixels(state);

    for (int kp = 0; kp < C8_ARRCOUNT(state->keypad.keys);
    kp++) {
        C8_Key * k = &(state->keypad.keys [ kp ]);
        c8_reset_key(k);
    }

    for (int ck = 0; ck < C8_ARRCOUNT(state->control_keys.keys);
    ck++) {
        C8_Key * k = &(state->control_keys.keys [ ck ]);
        c8_reset_key(k);
    }

    c8_reset_key(&(state->mouse_buttons.left_button));
    c8_reset_key(&(state->mouse_buttons.right_button));

    state->frame_count++;

    if (state->sound_timer > 0) {
        state->sound_timer--;
        state->should_beep = true;
    }
    else {
        state->should_beep = false;
    }

    if (state->delay_timer > 0) {
        state->delay_timer--;
    }
}


void c8_app_update(C8_State * state) {
    state->color_vertex_count = 0;
    state->text_vertex_count = 0;

    c8_update_emulator(state);
}


bool c8_arena_init(C8_Arena * arena, psz size, i32 alignement) {
    bool result = false;

    assert(size % alignement == 0);

    void *mem = c8_allocate(size);

    if (mem != 0) {
        arena->max_bytes = size;
        arena->alignement = alignement;
        arena->offset = 0;
        arena->data = mem;
        result = true;
    }
    else {
        C8_LOG_ERROR("Arena initialization failed");
    }

    return result;
}


void *c8_arena_alloc(C8_Arena * arena, psz size) {
    assert(arena != 0);

    void *result = 0;

    psz eff_size = size;
    if (eff_size % arena->alignement != 0) {
        eff_size += arena->alignement - (eff_size % arena->alignement);
    }

    bool big_enough = arena->offset + eff_size <= arena->max_bytes;

    assert(big_enough);

    if (big_enough) {
        result = ((u8 *) arena->data) + arena->offset;
        arena->offset += eff_size;
    }

    return result;
}


void c8_arena_free_all(C8_Arena * arena) {
    assert(arena != 0);

    arena->offset = 0;
}


#endif // !CHIP8_C\

