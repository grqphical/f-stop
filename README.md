# f-stop
A self-hosted, Google Photos-like web application that organizes your photos with a unique tag system.

Rather than just having folders for all your photos, you assign each photo tags that can then be used to filter through your photo library.

## Usage

### Dependencies
f-stop requires [libvips](https://www.libvips.org/install.html) and pkg-config to be installed.

### Installation
```bash
git clone https://github.com/grqphical/f-stop
cd f-stop/
go build -o f-stop ./cmd/api
```

## License
f-stop is licensed under the [Apache 2.0 License](LICENSE). f-stop uses [libvips](https://www.libvips.org/) which is licensed under the [GNU Lesser General Public License v2.1 or later](https://spdx.org/licenses/LGPL-2.1-or-later).
