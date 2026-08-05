# Learnings

## Security error handling pattern (confirmed in PR #596)

For error paths that touch user-supplied input (headers, query params, body fields):
1. Error values returned from validators must use static strings -- never embed user input via fmt.Errorf
2. HTTP response bodies must use static messages -- never write err.Error() when the error may contain user input
3. Test assertions should include negative checks: `ToNot(ContainSubstring(inputValue))` to verify no leakage
