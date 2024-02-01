# weather-api

A simple Go app for searching and displaying weather.

## Usage

This project is most easily run with `docker compose` as it consists of a Go
application and a database.

First, create a file called `.env` based on the values in `template.env`. Fill
in relevant values.

Next, simply run the following command to build and run the app:

```sh
docker compose up --build
```

Now, visit `localhost:8080` to test it out.

## Benchmarking

For HTTP load testing, [drill](https://github.com/fcsonline/drill) can be used.
See their documentation for installation.

Once installed, simply run the following to see how the API responds to large
numbers of concurrent requests:

```sh
drill --benchmark benchmark.yml --stats
```

This should generate a performance report something as follows:

```sh
Concurrency 20
Iterations 20
Rampup 0
Base URL http://localhost:8080

Fetch some cities from CSV http://localhost:8080/weather?city=Oslo 200 OK 136ms
Fetch some cities from CSV http://localhost:8080/weather?city=Oslo 200 OK 186ms
...

Fetch some cities from CSV Total requests            280
Fetch some cities from CSV Successful requests       280
Fetch some cities from CSV Failed requests           0
Fetch some cities from CSV Median time per request   143ms
Fetch some cities from CSV Average time per request  153ms
Fetch some cities from CSV Sample standard deviation 41ms
Fetch some cities from CSV 99.0th percentile        281ms
Fetch some cities from CSV 99.5th percentile        287ms
Fetch some cities from CSV 99.9th percentile        295ms

Time taken for tests      2.3 seconds
Total requests            280
Successful requests       280
Failed requests           0
Requests per second       122.21 [#/sec]
Median time per request   143ms
Average time per request  153ms
Sample standard deviation 41ms
99.0th percentile        281ms
99.5th percentile        287ms
99.9th percentile        295ms
```
