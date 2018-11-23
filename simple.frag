#version 330

uniform vec3 color;
uniform bool useExternalColor;

in fData
{
    vec3 color;
} g_in;


out vec4 colorOut;

void main() {

    if (useExternalColor) {
        colorOut = vec4(color,1);
    } else {
        colorOut = vec4(g_in.color, 1);
    }

}
