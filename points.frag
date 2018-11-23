#version 330

uniform vec3 color;
uniform sampler2D imageTexture;
uniform bool useExternalColor;

in vec2 vUV;
out vec4 colorOut;

void main() {


    vec2 coord = gl_PointCoord - vec2(0.5);  //from [0,1] to [-0.5,0.5]

    if(length(coord) > 0.5) {
        discard;
    }

    if (useExternalColor) {
        colorOut = vec4(color,1);
    } else {
        colorOut = texture(imageTexture, vUV);
    }


}
