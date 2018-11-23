#version 330

layout (location = 0) in vec3 vertPos;
layout (location = 1) in vec3 vertNormal;
layout (location = 2) in vec2 vertUV;

uniform mat4 viewProjectionMat;
uniform mat4 modelMat;

out vec2 vUV;

void main() {
    gl_Position = viewProjectionMat * modelMat * vec4(vertPos,1);
    vUV = vec2(vertUV.x, 1.0-vertUV.y);
}
