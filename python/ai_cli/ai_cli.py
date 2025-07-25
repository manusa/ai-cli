import os
import platform
import subprocess
import sys
from pathlib import Path
import shutil
import tempfile
import urllib.request

if sys.version_info >= (3, 8):
    from importlib.metadata import version
else:
    from importlib_metadata import version

__version__ = version("python-ai-cli")

def get_platform_binary():
    """Determine the correct binary for the current platform."""
    system = platform.system().lower()
    arch = platform.machine().lower()

    # Normalize architecture names
    if arch in ["x86_64", "amd64"]:
        arch = "amd64"
    elif arch in ["arm64", "aarch64"]:
        arch = "arm64"
    else:
        raise RuntimeError(f"Unsupported architecture: {arch}")

    if system == "darwin":
        return f"ai-cli-darwin-{arch}"
    elif system == "linux":
        return f"ai-cli-linux-{arch}"
    elif system == "windows":
        return f"ai-cli-windows-{arch}.exe"
    else:
        raise RuntimeError(f"Unsupported operating system: {system}")

def download_binary(binary_version="latest", destination=None):
    """Download the correct binary for the current platform."""
    binary_name = get_platform_binary()
    if destination is None:
        destination = Path.home() / ".ai-cli" / "bin" / binary_version

    destination = Path(destination)
    destination.mkdir(parents=True, exist_ok=True)
    binary_path = destination / binary_name

    if binary_path.exists():
        return binary_path

    base_url = "https://github.com/manusa/ai-cli/releases"
    if binary_version == "latest":
        release_url = f"{base_url}/latest/download/{binary_name}"
    else:
        release_url = f"{base_url}/download/v{binary_version}/{binary_name}"

    # Download the binary
    print(f"Downloading {binary_name} from {release_url}")
    with tempfile.NamedTemporaryFile(delete=False) as temp_file:
        try:
            with urllib.request.urlopen(release_url) as response:
                shutil.copyfileobj(response, temp_file)
            temp_file.close()

            # Move to destination and make executable
            shutil.move(temp_file.name, binary_path)
            binary_path.chmod(binary_path.stat().st_mode | 0o755)  # Make executable

            return binary_path
        except Exception as e:
            os.unlink(temp_file.name)
            raise RuntimeError(f"Failed to download binary: {e}")

def execute(args=None):
    """Download and execute the ai-cli binary."""
    if args is None:
        args = []

    try:
        binary_path = download_binary(binary_version=__version__)
        cmd = [str(binary_path)] + args

        # Execute the binary with the provided arguments
        process = subprocess.run(cmd)
        return process.returncode
    except Exception as e:
        print(f"Error executing ai-cli: {e}", file=sys.stderr)
        return 1

if __name__ == "__main__":
    sys.exit(execute(sys.argv[1:]))


def main():
    """Main function to execute the ai-cli binary."""
    args = sys.argv[1:] if len(sys.argv) > 1 else []
    return execute(args)
