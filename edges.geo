#version 330

layout (lines) in;
layout (line_strip, max_vertices = 2) out;

uniform sampler2D imageTexture;

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
        averageColor += texture(imageTexture, v_in[i].uv).rgb;
    }
    averageColor /= float(gl_in.length());

    for(i = 0; i < gl_in.length(); i++) {
        gl_Position = gl_in[i].gl_Position;

        g_out.color = averageColor;

        EmitVertex();
    }
    EndPrimitive();

}
