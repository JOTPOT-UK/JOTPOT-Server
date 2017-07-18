if [ ! -d "Release" ]
then
	mkdir Release
fi

cd jps
tar -cf ../Release/jps.tar *
tar -czf ../Release/jps.tar.gz *
zip -r ../Release/jps.zip *
cd ..

cd jps-source
tar -cf ../Release/jps-source.tar *
tar -czf ../Release/jps-source.tar.gz *
zip -r ../Release/jps-source.zip *
cd ..

cd build/source/build
tar -cf ../../../Release/jps-build-ready.tar *
tar -czf ../../../Release/jps-build-ready.tar.gz *
zip -r ../../../Release/jps-build-ready.zip *
cd ..
