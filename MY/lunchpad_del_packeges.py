from launchpadlib.launchpad import Launchpad
from packaging.version import Version
import os

cachedir = os.path.expanduser("~/.launchpadlib/cache")
launchpad = Launchpad.login_with("delete-script", "production", cachedir, version="devel")
ppa = launchpad.people["boosterykt"].getPPAByName(name="wal-g")

for pub in ppa.getPublishedSources(status="Superseded"):
    ver = pub.source_package_version
    try:
        if Version(ver.split("-")[0]) < Version("18.0.0"):
            print(f"Requesting deletion of {ver}...")
            pub.requestDeletion(removal_comment="Cleaning up old versions")
    except Exception as e:
        print(f"Skipping {ver}: {e}")