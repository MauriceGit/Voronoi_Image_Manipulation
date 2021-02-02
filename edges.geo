#version 330

layout (lines) in;
layout (line_strip, max_vertices = 2) out;

uniform sampler2D imageTexture;
uniform vec3 color;
uniform bool useExternalColor;

in vData
{
    vec2 uv;
} v_in[];

out fData
{
    vec3 color;
} g_out;

void main() {

    int i;
    vec3 averageColor = vec3(0,0,0);
    for(i = 0; i < gl_in.length(); i++) {
        vec4 tex = texture(imageTexture, v_in[i].uv).rgba;
        if (useExternalColor) {
            averageColor += mix(vec3(0), color, tex.a);
        } else {
            averageColor += mix(vec3(0), tex.rgb, tex.a);
        }
    }
    averageColor /= float(gl_in.length());

    for(i = 0; i < gl_in.length(); i++) {
        gl_Position = gl_in[i].gl_Position;

        g_out.color = averageColor;

        EmitVertex();
    }
    EndPrimitive();

}
