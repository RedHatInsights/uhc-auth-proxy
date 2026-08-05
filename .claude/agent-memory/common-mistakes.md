# Common Mistakes

## Partial vulnerability fixes

When remediating an information disclosure or injection vulnerability, grep the entire codebase for all instances of the vulnerable pattern before submitting the fix. A partial fix that leaves other call sites exposed will be caught in review and delays merge.
