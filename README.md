# Fetch_And_Go
## Setup
1. Ensure the `Go` programming language is installed on your system.
   You can use your system package manager (apt, pacman, homebrew, ...) or you can follow [https://go.dev/doc/install](https://go.dev/doc/install). Make sure `go version` runs successfully.
   The code was tested with `go 1.20.1`, but earlier versions such as 1.18 should also run.
   
2. Clone the repository with `git clone https://github.com/wiscous/Fetch_And_Go`.

3. `cd` into the project directory with `cd Fetch_And_Go`;

4. You can now run the code with the sample input using `go run . 5000 ./transactions.csv` and get the output.

```
{
        "DANNON": 1000,
        "MILLER COORS": 5300,
        "UNILEVER": 0
}
```

