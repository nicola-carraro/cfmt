int main(void) {
    state->load_button.is_mouse_over = state->mouse_position.x >= load_button.x && state->mouse_position.x <= load_button.x + load_button.width
        && state->mouse_position.y >= load_button.y && state->mouse_position.y <= load_button.y + load_button.height;
}
