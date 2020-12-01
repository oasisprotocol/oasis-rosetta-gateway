# Punch configuration file.

# For more information, see: https://punch.readthedocs.io/.

__config_version__ = 1

GLOBALS = {
    'serializer': {
        'semver': '{{ major }}.{{ minor }}.{{ patch }}',
     }
}

# NOTE: The FILES list is not allowed to be empty, so we need to pass it at
# least a single valid file.
FILES = ["README.md"]

VERSION = ['major', 'minor', 'patch']
