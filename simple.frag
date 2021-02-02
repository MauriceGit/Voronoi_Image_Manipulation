#version 330

in fData
{
    vec3 color;
} g_in;

out vec4 colorOut;

void main() {
    colorOut = vec4(g_in.color, 1);
}
