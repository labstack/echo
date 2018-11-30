tag:
	@git tag `grep -P '^\tversion = ' echo.go|cut -f2 -d'"'`
	@git tag|grep -v ^v
