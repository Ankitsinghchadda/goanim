// Package icons is goanim's curated set of generic system-design
// icons (Client, Server, Database, Queue, Stack, Cache, LoadBalancer,
// MessageBroker, Worker, APIGateway, CDN, User, Service, Function,
// Storage).
//
// Each icon is a composition of style-aware primitives (rectangles,
// ellipses, lines, custom paths). The icon's *geometry* is fixed by
// the type (a queue is recognizably a queue); its *rendering* picks up
// the active style — rough in Sketchy mode, clean in Crisp mode, etc.
//
// All icons satisfy the icon.Icon and mobject.Attachable interfaces,
// so arrows route to them naturally. The default label position is
// LabelBelow; override with WithLabelPosition.
package icons
