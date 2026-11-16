import type { Route } from "@/lib/route";
import Top from "@/views/Top";
import Trends from "@/views/Trends";
import Week from "@/views/Week";

/** Projects is where the time went: the week's or the month's charts, or
 *  the ranking for a day. */
export default function Projects({ route, tick }: { route: Route; tick: number }) {
  if (route.unit === "week") return <Week date={route.date} tick={tick} />;
  if (route.unit === "month") return <Trends date={route.date} />;
  return <Top tick={tick} />;
}
