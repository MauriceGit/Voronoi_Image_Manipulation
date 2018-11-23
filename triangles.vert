#version 330

layout (location = 0) in vec3 vertPos;
layout (location = 1) in vec3 vertNormal;
layout (location = 2) in vec2 vertUV;

uniform mat4 viewProjectionMat;
uniform mat4 modelMat;

uniform float expectedRadiusX;
uniform float expectedRadiusY;

uniform sampler2D imageTexture;

out fData
{
    vec3 color;
} g_out;

// Samples randomly within the radius r around pos and returns the average sample color over all 6 samples.
vec3 sampleTextureRandom(vec2 r, vec2 pos) {
    vec3 color = texture(imageTexture, pos + r*(2*vec2(0.36123,0.83771)-1.0)).rgb;
    color += texture(imageTexture, pos +     r*(2*vec2(0.47154,0.44896)-1.0)).rgb;
    color += texture(imageTexture, pos +     r*(2*vec2(0.93110,0.64977)-1.0)).rgb;
    color += texture(imageTexture, pos +     r*(2*vec2(0.15231,0.46326)-1.0)).rgb;
    color += texture(imageTexture, pos +     r*(2*vec2(0.83720,0.11699)-1.0)).rgb;
    color += texture(imageTexture, pos +     r*(2*vec2(0.30478,0.06818)-1.0)).rgb;

    return color/6;
}


void main() {
    gl_Position = viewProjectionMat * modelMat * vec4(vertPos,1);

    g_out.color = sampleTextureRandom(vec2(expectedRadiusX, expectedRadiusY), vec2(vertUV.x, 1.0-vertUV.y));
}
