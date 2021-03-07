# Bloom Filter in Ethereum

Bloom Filter 是一种可以快速检索的工具。Bloom Filter 本身由是一个长度为m的bit array，k个不相同的hash 函数和源dataset组成。具体的说，Bloom Filter是由k个不同的hash function将源dataset hash到m位的bit array构成。通过Bloom Filter，我们可以快速检测出，一个data是不是在源dataset中（O(k) time）。

Bloom Filter不保证完全的正确性。In other word, Bloom Filter 会出现false positive的结果。但是，如果一个被检索的data，在bloom filter 中得到了false的反馈，那他一定不在源data之中。

Bloom Filter是Ethereum的核心功能之一。具体的代码是实现是在“core/types/bloom9.go”文件之中。

在文件的起始位置，定义了两个常量BloomByteLength 和 BloomBitLength，大小分别是256和8.

在Ethereum中的Bloom Filter是一个长度为256的byte数组组成的。

"""
    Bloom represents a 2048 bit bloom filter
    type Bloom [BloomByteLength]byte
"""