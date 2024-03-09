#ifdef _MSC_VER
#ifdef __cplusplus
extern "C" {
#endif

	void _rt0_amd64_windows_lib();

	__pragma(section(".CRT$XCU", read));
	__declspec(allocate(".CRT$XCU")) void (*init1)() = _rt0_amd64_windows_lib;
	__pragma(comment(linker, "/include:init1"));

#ifdef __cplusplus
}
#endif
#endif

int main() { return 0; }
