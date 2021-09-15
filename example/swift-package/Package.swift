// swift-tools-version:5.3
import PackageDescription

let package = Package(
	name: "GetKit",
	products: [
		.library(
			name: "GetKit",
			targets: ["GetGo"]
		),
	],
	targets: [
		.target(
			name: "GetKit"
		),
		.binaryTarget(
			name: "GetGo",
			path: "Frameworks/GetGo.xcframework"
		),
		.testTarget(
			name: "GetKitTests",
			dependencies: ["GetKit", "GetGo"]
		),
	]
)
