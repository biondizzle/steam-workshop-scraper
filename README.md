Steam's Workshop Search is broken.

I couldn't find any way to query it through thier API unless I am the developer of the game

But the docs are a mess, maybe I'm missing it?

This is just a quick way to parse the pages

TODO:
  - Fix `published_at_unix` field. I stupidly misread `data-publishedfileid` as `"DATE"-publishedfileid`. So just needs to be grabbed from the Individuakl Workshop Page inside the div `.detailsStatsContainerRight`

  - Increment `Current Items` when `WRITE_DIRECT_TO_DB` is true just for some output on the screen

 - Make everything configurable from the `settings.json`

 - Release a build