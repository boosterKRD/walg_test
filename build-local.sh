          # Create rules file
          echo "#!/usr/bin/make -f" > debian/rules
          echo "export DH_GOPKG=${DH_GOPKG}" >> debian/rules
          echo "%" >> debian/rules
          printf "\tdh \$@ --buildsystem=golang\n" >> debian/rules  # Используем printf вместо echo
          chmod +x debian/rules
