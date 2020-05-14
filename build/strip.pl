#!/usr/bin/env perl
use strict;
use warnings;


print STDERR $ARGV[0],$/;

my $count = 0;

binmode STDIN, ":utf8";
binmode STDOUT, ":utf8";

while ($_ = <STDIN>) {
	$count++;
	next if length($_) > $ARGV[0];
	print $_;

}
print STDERR "$count rows\n";
