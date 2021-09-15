import XCTest

@testable import GetGo

class GetTests: XCTestCase {
	func testGet() {
		var error: NSErrorPointer = nil
		let response = GoGet("https://golang.org/", error)
		XCTAssertNil(error)
		guard let response = response else {
			XCTFail("response == nil")
			return
		}
		guard let str = String(data: response, encoding: .utf8) else {
			XCTFail("str == nil")
			return
		}
		XCTAssert(str.contains("Go"))
		XCTAssert(str.contains("an open source programming language"))
		XCTAssert(str.contains("https://play.golang.org"))
	}

	func testVersion() {
		let version = GoVersion()
		XCTAssertEqual(version, "0.0.1")
	}
}
