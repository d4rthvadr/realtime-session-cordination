"use client";

import * as React from "react";
import * as RechartsPrimitive from "recharts";

import { cn } from "@/lib/utils";

export type ChartConfig = {
  [k: string]: {
    label?: React.ReactNode;
    color?: string;
  };
};

type ChartContextProps = {
  config: ChartConfig;
};

const ChartContext = React.createContext<ChartContextProps | null>(null);

function useChart() {
  const context = React.useContext(ChartContext);
  if (!context) {
    throw new Error("useChart must be used within a <ChartContainer />");
  }
  return context;
}

export const ChartContainer = React.forwardRef<
  HTMLDivElement,
  React.ComponentProps<"div"> & {
    config: ChartConfig;
    children: React.ComponentProps<
      typeof RechartsPrimitive.ResponsiveContainer
    >["children"];
  }
>(({ id, className, children, config, ...props }, ref) => {
  const chartId = React.useId();
  const resolvedId = `chart-${id || chartId.replace(/:/g, "")}`;

  return (
    <ChartContext.Provider value={{ config }}>
      <div
        data-chart={resolvedId}
        ref={ref}
        className={cn(
          "flex aspect-video justify-center text-xs [&_.recharts-cartesian-axis-tick_text]:fill-muted-foreground [&_.recharts-cartesian-grid_line[stroke='#ccc']]:stroke-border [&_.recharts-curve.recharts-tooltip-cursor]:stroke-border [&_.recharts-pie-label-text]:fill-foreground [&_.recharts-polar-grid_[stroke='#ccc']]:stroke-border [&_.recharts-radial-bar-background-sector]:fill-muted [&_.recharts-rectangle.recharts-tooltip-cursor]:fill-muted [&_.recharts-reference-line_[stroke='#ccc']]:stroke-border [&_.recharts-sector[stroke='#fff']]:stroke-transparent [&_.recharts-sector]:outline-none",
          className,
        )}
        {...props}
      >
        <ChartStyle id={resolvedId} config={config} />
        <RechartsPrimitive.ResponsiveContainer>
          {children}
        </RechartsPrimitive.ResponsiveContainer>
      </div>
    </ChartContext.Provider>
  );
});
ChartContainer.displayName = "ChartContainer";

const ChartStyle = ({ id, config }: { id: string; config: ChartConfig }) => {
  const colorConfig = Object.entries(config).filter(([, conf]) => conf.color);
  if (!colorConfig.length) {
    return null;
  }

  return (
    <style
      dangerouslySetInnerHTML={{
        __html: `
[data-chart=${id}] {
${colorConfig
  .map(([key, itemConfig]) => `  --color-${key}: ${itemConfig.color};`)
  .join("\n")}
}
`,
      }}
    />
  );
};

export const ChartTooltip = RechartsPrimitive.Tooltip;

export function ChartTooltipContent({
  active,
  payload,
  className,
  hideLabel = false,
}: {
  active?: boolean;
  payload?: Array<{
    dataKey?: string | number;
    name?: string | number;
    value?: unknown;
  }>;
  className?: string;
  hideLabel?: boolean;
}) {
  const { config } = useChart();

  if (!active || !payload || !payload.length) {
    return null;
  }

  return (
    <div
      className={cn(
        "grid min-w-[8rem] gap-1 rounded-lg border bg-background px-2.5 py-1.5 text-xs shadow-xl",
        className,
      )}
    >
      {payload.map((item, index) => {
        const key = String(item.dataKey ?? item.name ?? "value");
        const label = config[key]?.label ?? item.name ?? key;

        return (
          <div
            key={`${key}-${index}`}
            className="flex items-center justify-between gap-2"
          >
            {!hideLabel && (
              <span className="text-muted-foreground">{label}</span>
            )}
            <span className="font-mono font-medium text-foreground">
              {item.value as React.ReactNode}
            </span>
          </div>
        );
      })}
    </div>
  );
}
