# Go Dimensional Flag-Fields Experiment

The below documents an experiment I performed in creation of a library for modelling of dimensional flag fields.

# Hypothesis

There are a class of problems that can be solved as operations on a multi-dimensional series of [bit fields](https://en.wikipedia.org/wiki/Bit_field). For the remainder of this readme we will specialize the term "bit-field" to "flag-field" where every bit represents a boolean or flag. 

Doing so allows use to compact more information into a smaller space. This makes it more likely we're going to hit the [CPU cache](https://en.wikipedia.org/wiki/CPU_cache). This is important [as latency increases with cache misses](http://highscalability.com/blog/2013/1/15/more-numbers-every-awesome-programmer-must-know.html).

At present [each boolean in Golang takes up a single byte](https://medium.com/@val_deleplace/cost-of-a-bool-slice-in-go-b7d7ba1d6dda) meaning that to store a single bits worth of information we store an additional 7. This has an impact on cache use efficiency.

By wrapping this knowledge into a library we can allow others to easily have the same cache locality performance gainz (sic).

## Motivating Example 
Consider the example of a simple 2D maze. For each cell we can specify which directions we can go in and whether or not the given cell is "special", say as it is the entry into the maze, with a simple flag-field where each bit describes the call:

```
|---- Unused as Go is byte addressable  
| |   ( more on on this later)
| |  |- 1 means we can go north here
| |  | |- 1 means we can go south here
00010101
   || |
   || |- 0 means we can't go west here
   ||- 0 means we we can't go east here
   |- 1 means this cell is the entry point
```

We could represent the above as a series of 5 booleans. [Each boolean in Golang takes up a single byte](https://medium.com/@val_deleplace/cost-of-a-bool-slice-in-go-b7d7ba1d6dda). This means to represent a 10 x 10 grid we'll use:

```
10 rows * 10 columns * 5 bytes = 500 bytes
```

Assuming we use a byte with the flag-field based approach, accepting that we'll have 3 bits unused, this reduces to 

```
10 rows * 10 columns * 1 byte = 100 bytes
```

If we want to get really smart and not accept 3 bits of fragmentation this can reduce to :

```
(10 rows * 10 columns * 5 bits) / 8 bits to a byte = 63 bytes
```
Meaning we could in theory [store our entire maze in a single 64 byte cache line](https://www.sciencedirect.com/topics/computer-science/cache-line-size).

For a 10x10 maze there likely isn't much difference in performance as we're either using one cache line or two which are both likely to be present in the cache. However as we scale to higher dimensions the number of cache lines requires increases: 

| Approach | Grid dimension | Bytes required | Cache lines required |
| - | - | - | - |
| Series of bools | 100  | 50000  | 781  |
| Byte field | 100  | 10000  | 156  |
| Packed fields | 100  | 6250  | 97  |

By using packed flag fields here we have a higher chance that the data required will be present in the cache. This is important when considering the cache may be washed by other processes.

## What are the assumptions / constraints on a possible solution? 

Below are a set of assumptions put on the solution. How true these assumptions are depends on your use case.

1. Typically the cardinality of a flag-field is known up front: we know how many flags we want at compile time. In terms of the size of cardinality we're likely to see smaller values of < 10.
3. If we were to stick this in a library it should be generic and not prescriptive to the problems seen thus far. The counter balance here is that the abstractions need to be as cheap / fast as possible.
4. In the worst case the user of the library will be selecting fields completely at random. In practice it's more likely that fields will be selected close to other recently selected fields. In the maze example above we're likely to grab the fields for the cells above, below and next to the current cell we care about soon after the current cell. However with close together field accesses we're likely leveraging some cache locality regardless of underlying approach. Therefore test with random accesses of random fields.
5. Concurrent modification across multiple threads / goroutines isn't a priority / is likely killing performance anyway. 
6. Extra CPU Instructions are acceptable, within reason, if we see speedup on a large number of elements.

### Motivation for exploration

As part of learning [Golang](https://go.dev/) I did a number of practice programming tasks. A fair number of them involved operating on some kind of grid data structure then doing some computation. 

Having an interest in efficiency led me to certain subjects such as [data orientated design](https://www.youtube.com/watch?v=rX0ItVEVjHc), [mechanical sympathy](https://wa.aws.amazon.com/wellarchitected/2020-07-02T19-33-23/wat.concept.mechanical-sympathy.en.html#:~:text=Mechanical%20sympathy%20is%20when%20you,have%20to%20have%20Mechanical%20Sympathy.) and ultimately [how to best leverage the cache](https://www.youtube.com/watch?v=WDIkqP4JbkE). This influenced the design of the solutions I put in place. 

After re-creating the same approach several times for differing tasks I figured a general approach could be put together and tested for speed.

# Experiment

Prior to creating a 2 dimensional version of this I initially produced a [1 dimensional version](./one_dimensional.go) to validate the potential speed up. Behind the scenes it packs the N cardinality flag fields into a series of integers in a slice. 

To perform modification / reading of any fields it must first determine which element of the underlying slice the field starts with then determine which bits of the field require reading. This represents a bit of CPU work ahead of time that have been optimized where possible.

As well as validating the implementation worked I benchmarked it against [a series of other approaches](./one_dimensional_test.go) that are detailed below:

|name|description|
|-  |    -|
|Arrays| Represent each flag-field as a `[constant]bool` with the index going into the array|
|Slices| Represent each flag-field as a `[]bool` with the index going into the slice|
|Map| Represent each flag-field as a `map[byte]bool` with the key being the field to lookup|
|Struct| Have each field being a `Val bool` field in a struct definition. *Note that for the benchmark below we still lookup fields numerically. This means we have to translate from a number to a field. This is likely quicker with straight field access*|
|Bitwise| Having appropriately sized int types and using bit masks to set / fetch the relevant fields|

For each of the above and the library implementation the psuedo code for the benchmark was :

```
  for i := 0; i < testRunCount; i++ {
     grab a random flag-field and set a random field on it to 1 
     grab a random flag-field and set a random field on it to 0 (a.k.a unset)
     grab a random flag-field and get a random field 
  }
```

The library implementation assumes you would like to operate on multiple fields within a flag-field at a given time. However the other implementations think in terms of single field operations. To see if the multiple field operations were a cause of slowdown I also created simple "single field" operations on the library implementation that is also measured.

# Results

To validate the approach vs others the standard [Go Benchmarking package](https://pkg.go.dev/testing#hdr-Benchmarks) was used and executed with:

```
go test -bench=. -benchmem -cpuprofile profile.out
```

The results below are consistent across multiple runs.
|Approach|Element count|Scenario| ns/op|
| -     | - | -              |    -    |
|Library|100|Random_access-16|0.0003329|
|Library_Single_access|100|Random_access-16|0.0003050|
|Arrays|100|Random_access-16|0.0002714|
|Slices|100|Random_access-16|0.0002615|
|Bitwise16Bit|100|Random_access-16|0.0002658|
|Bitwise32Bit|100|Random_access-16|0.0002692|
|Structs|100|Random_access-16|0.0004498|
|Maps|100|Random_access-16|0.0007946|
|Library|1000|Random_access-16|0.0003230|
|Library_Single_access|1000|Random_access-16|0.0002943|
|Arrays|1000|Random_access-16|0.0002566|
|Slices|1000|Random_access-16|0.0002752|
|Bitwise16Bit|1000|Random_access-16|0.0002650|
|Bitwise32Bit|1000|Random_access-16|0.0002630|
|Structs|1000|Random_access-16|0.0004577|
|Maps|1000|Random_access-16|0.0008367|
|Library|10000|Random_access-16|0.0003110|
|Library_Single_access|10000|Random_access-16|0.0003138|
|Arrays|10000|Random_access-16|0.0002675|
|Slices|10000|Random_access-16|0.0002716|
|Bitwise16Bit|10000|Random_access-16|0.0002532|
|Bitwise32Bit|10000|Random_access-16|0.0002547|
|Structs|10000|Random_access-16|0.0004425|
|Maps|10000|Random_access-16|0.001198|
|Library|100000|Random_access-16|0.0003025|
|Library_Single_access|100000|Random_access-16|0.0002930|
|Arrays|100000|Random_access-16|0.0002610|
|Slices|100000|Random_access-16|0.0002976|
|Bitwise16Bit|100000|Random_access-16|0.0002571|
|Bitwise32Bit|100000|Random_access-16|0.0002555|
|Structs|100000|Random_access-16|0.0004723|
|Maps|100000|Random_access-16|0.001741|
|Library|1000000|Random_access-16|0.0004384|
|Library_Single_access|1000000|Random_access-16|0.0003563|
|Arrays|1000000|Random_access-16|0.001021|
|Slices|1000000|Random_access-16|0.001230|
|Bitwise16Bit|1000000|Random_access-16|0.0004016|
|Bitwise32Bit|1000000|Random_access-16|0.0005177|
|Structs|1000000|Random_access-16|0.0008266|
|Maps|1000000|Random_access-16|0.004522|
|Library|10000000|Random_access-16|0.001354|
|Library_Single_access|10000000|Random_access-16|0.001340|
|Arrays|10000000|Random_access-16|0.001096|
|Slices|10000000|Random_access-16|0.001497|
|Bitwise16Bit|10000000|Random_access-16|0.001109|
|Bitwise32Bit|10000000|Random_access-16|0.001169|
|Structs|10000000|Random_access-16|0.0009460|
|Maps|10000000|Random_access-16|0.006060|
|Library|10000000|Random_access#01-16|0.001340|
|Library_Single_access|10000000|Random_access#01-16|0.001345|
|Arrays|10000000|Random_access#01-16|0.001086|
|Slices|10000000|Random_access#01-16|0.001463|
|Bitwise16Bit|10000000|Random_access#01-16|0.001132|
|Bitwise32Bit|10000000|Random_access#01-16|0.001162|
|Structs|10000000|Random_access#01-16|0.0009286|
|Maps|10000000|Random_access#01-16|0.005888|
|Library|100000000|Random_access-16|0.001912|
|Library_Single_access|100000000|Random_access-16|0.001909|
|Arrays|100000000|Random_access-16|0.005645|
|Slices|100000000|Random_access-16|0.005503|
|Bitwise16Bit|100000000|Random_access-16|0.001420|
|Bitwise32Bit|100000000|Random_access-16|0.002078|
|Structs|100000000|Random_access-16|0.001652|
|Maps|100000000|Random_access-16|0.01173|

Taken with the below from the benchmark output :
```
grep "BenchmarkOneDimenstionalVsAlternateApproachesForSingleFields" results | gsed -e "s/BenchmarkOneDimenstionalVsAlternateApproachesForSingleFields\///g;s/\s\+/ /g"  | cut -d " " -f1,3 | gsed -e "s/^\|$/|/g" | tr " " "|" 
```

# Analysis

From the above we can see that the static array and bitwise backed implementation outperform the library implementation before higher numbers of elements of flag fields.

The current theory as to why this is the case comes down to the number of extra CPU instructions required for the library implementation. 
The library implementation requires us to determine where a given field is some contiguous memory. Doing so requires instructions that are insanely cheap but represents more work than either the bitwise or array implementations. In fact some of the instructions may actually be handled implicitly by the underlying assembly language e.g. array lookups and stores.

While the array and bitwise implementations are less compact than the library implementation in memory they are still extremely cache friendly. 

This means that at lower number of elements you're likely dealing with cached values across all three implementations. It therefore comes down to number of CPU work required to access values. This is evidenced by the bitwise implementation being slower than the array implementation as it requires more instructions than array accesses while requiring less than the library implementation.

At a higher number of elements the advantage of the compactness of the library implementation starts to briefly shine through. As well as beating the array implementation it starts to win out against the wider bitwise implementations, touching on the narrower bitwise implementation. 

At the extreme end the randomness of the access pattern starts to become the deciding factor: subsequent runs yield differing results from one another. 

## What about the other approaches?

### Slices
A noticeable difference is that slices are slower than static arrays for this problem. This makes sense as slices have bounds checking in place which eats into CPU time. In addition the extra runtime information for length / cap affect cache locality as it increases the amount of space a sequential store takes up.

### Structs
The struct approach deserves a specific call out of a different approach that may yield faster results in the real world. While the results here are slower that is because we perform a translation from an index based lookup to a field of the struct to support random access. In the real world you likely want to look up properties that have a given name for example with the maze above you may represent the cell as 
```
MazeCell {
   CanGoLeft bool
   CanGoRight bool
   CanGoUp bool
   CanGoDown bool
   IsStart bool
}
```
with the idea being logic would interrogate each direction in sequence vs reading random fields in a random order.
This approach means the compiler will know the exact offsets from the start of a struct at compile time vs having to look things up in an array at runtime.

### Maps
The Map based approach is universally slower. Behind the scenes the memory benchmark suggests allocations occurring despiting creating each map to handle the required number of elements. With this in mind this approach can be discarded.

# Conclusions

While the library implementation yields better results for higher element counts the likelihood of having those larger element count in real life scenarios is questionable. These results would also manifest with a higher cardinality of flag fields. In practice it might be that the cache pressure presented with larger element counts is actually represented by other data structures or information used while processing the field sets requiring cache space.

With that said the recommendation from this experiment would be to try an array based implementation first if the number of flag fields is known ahead of time. In cases where the number of flag fields isn't known then a bitwise based approach (with appropriate selection of the underlying int type) is a decent alternative.

Only if memory is truly scarce and you're willing to pay for extra CPU time for it then consider the library approach. Alternatively if you really have a large number of flag fields, or a large number of high cardinality, then consider it as well.

# Reflections

Primarily I will say this is more an example or re-enforcement of well known [premature optimization is the root of all evil](https://wiki.c2.com/?PrematureOptimization):

1. Always measure up front to validate assumptions. Rather than creating a full blown implementation I opted to validate on the approach prior to committing.
2. I will happily hold my hand up to this being a pre-mature optimization. 
3. Removing work from the critical path tends to yield higher throughput / performance benefits.
