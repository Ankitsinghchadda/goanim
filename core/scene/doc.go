// Package scene hosts the Scene struct — the timeline, camera, and
// mobject set that drive video output. Scenes are composed by adding
// mobjects, then played with Scene.Play, which advances animations
// and streams frames to a render sink.
package scene
