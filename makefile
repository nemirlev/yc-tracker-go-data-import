.SILENT: package

package:
	echo "=========================================="
	echo "= Building zip archive for Go application ="
	echo "==========================================\n"
	rm -f build/tracker-data-import.zip \
    rm -f go.mod \
    mv go.serverles.mod go.mod \
    zip -r build/tracker-data-import.zip . \
    	-x config.yml -x deploy/\* -x docker-compose.yml -x go.sum
