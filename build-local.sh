          # Create rules file
          echo "#!/usr/bin/make -f" > debian/rules
          echo "export DH_GOPKG=${DH_GOPKG}" >> debian/rules
          echo "%" >> debian/rules
          echo -e "\tdh \$@ --buildsystem=golang" >> debian/rules
          chmod +x debian/rules
