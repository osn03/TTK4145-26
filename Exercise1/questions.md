Exercise 1 - Theory questions
-----------------------------

### Concepts

What is the difference between *concurrency* and *parallelism*?
> concurrrency er å gjøre flere oppgaver ved å hoppe mellom de, og dermed skape en illusjon om at de gjøres samtidig. Parallelism er å gjøre oppgavene samtidig.

What is the difference between a *race condition* and a *data race*? 
> Data race er at flere prosesser brukes samme minneadresse samtidig hvor minst en av de skriver til adressen. Race condition er en høynivå logikkfeil hvor timin eller rekkefølge påvirker riktigheten i koden. 
 
*Very* roughly - what does a *scheduler* do, and how does it do it?
> En scheduler håndterer hvilke prosesser som gjøres til enhver tid på dataens ulike ressurser.


### Engineering

Why would we use multiple threads? What kinds of problems do threads solve?
> Flere tråder kan utføre oppgaver raskere ved å  dele opp store oppgaver i flere deler og gjøre disse simultant.

Some languages support "fibers" (sometimes called "green threads") or "coroutines"? What are they, and why would we rather use them over threads?
> Fibers er er en form for oppdeling av oppgaver hvor man kan velge mellom flere arbeidsoppgaver på en enkelt tråd ved å starte og stoppe ulike fibre. De burde brukes der man skal utføre mange små oppgaver, men ikke kan tillate overhead kostnaden av flere tråder.

Does creating concurrent programs make the programmer's life easier? Harder? Maybe both?
> Vanskeligere fordi flere feil: race conditions, deadlock. Men avgjørende for å håndtere flere tråder.

What do you think is best - *shared variables* or *message passing*?
> Message passing er tryggere, men mer ressurskrevende. Shared variables er billigere og raskene, men feil kan oppstå lettere. Message passing er bedre der man har muligheten, og kan også brukes mellom flere ulike maskiner.


