
## Docker Image for gomobile

Build the Docker image with

    $ docker build -t gomobile .

Now run `gomobile bind` on your Go library code with

    $ docker run --rm -v $(pwd):/home/go/src/mylib/ \
        -w /home/go/src/mylib gomobile bind
