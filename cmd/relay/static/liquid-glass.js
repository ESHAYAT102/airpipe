(function () {
  const CONFIG = {
    radius: 0,
    depth: 6,
    blur: 6,
    strength: 100,
    chromaticAberration: 4,
  };

  const supportValue =
    "blur(1px) url('data:image/svg+xml,%3Csvg xmlns=%22http://www.w3.org/2000/svg%22/%3E#x')";

  function canUseLiquidGlass() {
    if (!window.CSS || !window.navigator) return false;

    const userAgent = window.navigator.userAgent.toLowerCase();
    const isFirefox = userAgent.includes("firefox");

    return (
      !isFirefox &&
      (CSS.supports("backdrop-filter", supportValue) ||
        CSS.supports("-webkit-backdrop-filter", supportValue))
    );
  }

  function getDisplacementMap({ height, width, radius, depth }) {
    return (
      "data:image/svg+xml;utf8," +
      encodeURIComponent(`<svg height="${height}" width="${width}" viewBox="0 0 ${width} ${height}" xmlns="http://www.w3.org/2000/svg">
    <style>
        .mix { mix-blend-mode: screen; }
    </style>
    <defs>
        <linearGradient
          id="Y"
          x1="0"
          x2="0"
          y1="${Math.ceil((radius / height) * 15)}%"
          y2="${Math.floor(100 - (radius / height) * 15)}%">
            <stop offset="0%" stop-color="#0F0" />
            <stop offset="100%" stop-color="#000" />
        </linearGradient>
        <linearGradient
          id="X"
          x1="${Math.ceil((radius / width) * 15)}%"
          x2="${Math.floor(100 - (radius / width) * 15)}%"
          y1="0"
          y2="0">
            <stop offset="0%" stop-color="#F00" />
            <stop offset="100%" stop-color="#000" />
        </linearGradient>
    </defs>

    <rect x="0" y="0" height="${height}" width="${width}" fill="#808080" />
    <g filter="blur(2px)">
      <rect x="0" y="0" height="${height}" width="${width}" fill="#000080" />
      <rect
          x="0"
          y="0"
          height="${height}"
          width="${width}"
          fill="url(#Y)"
          class="mix"
      />
      <rect
          x="0"
          y="0"
          height="${height}"
          width="${width}"
          fill="url(#X)"
          class="mix"
      />
      <rect
          x="${depth}"
          y="${depth}"
          height="${height - 2 * depth}"
          width="${width - 2 * depth}"
          fill="#808080"
          rx="${radius}"
          ry="${radius}"
          filter="blur(${depth}px)"
      />
    </g>
</svg>`)
    );
  }

  function getDisplacementFilter({
    height,
    width,
    radius,
    depth,
    strength = 100,
    chromaticAberration = 0,
  }) {
    return (
      "data:image/svg+xml;utf8," +
      encodeURIComponent(`<svg height="${height}" width="${width}" viewBox="0 0 ${width} ${height}" xmlns="http://www.w3.org/2000/svg">
    <defs>
        <filter id="displace" color-interpolation-filters="sRGB">
            <feImage x="0" y="0" height="${height}" width="${width}" href="${getDisplacementMap(
              {
                height,
                width,
                radius,
                depth,
              }
            )}" result="displacementMap" />
            <feDisplacementMap
                transform-origin="center"
                in="SourceGraphic"
                in2="displacementMap"
                scale="${strength + chromaticAberration * 2}"
                xChannelSelector="R"
                yChannelSelector="G"
            />
            <feColorMatrix
            type="matrix"
            values="1 0 0 0 0
                    0 0 0 0 0
                    0 0 0 0 0
                    0 0 0 1 0"
            result="displacedR"
                    />
            <feDisplacementMap
                in="SourceGraphic"
                in2="displacementMap"
                scale="${strength + chromaticAberration}"
                xChannelSelector="R"
                yChannelSelector="G"
            />
            <feColorMatrix
            type="matrix"
            values="0 0 0 0 0
                    0 1 0 0 0
                    0 0 0 0 0
                    0 0 0 1 0"
            result="displacedG"
                    />
            <feDisplacementMap
                    in="SourceGraphic"
                    in2="displacementMap"
                    scale="${strength}"
                    xChannelSelector="R"
                    yChannelSelector="G"
                />
                <feColorMatrix
                type="matrix"
                values="0 0 0 0 0
                        0 0 0 0 0
                        0 0 1 0 0
                        0 0 0 1 0"
                result="displacedB"
                        />
              <feBlend in="displacedR" in2="displacedG" mode="screen"/>
              <feBlend in2="displacedB" mode="screen"/>
        </filter>
    </defs>
</svg>`) +
      "#displace"
    );
  }

  function applyLiquidGlass(nav) {
    const existingLayer = nav.querySelector(":scope > .liquid-glass-layer");
    const layer = existingLayer || document.createElement("div");

    if (!existingLayer) {
      layer.className = "liquid-glass-layer";
      layer.setAttribute("aria-hidden", "true");
      nav.prepend(layer);
    }

    const supportsLiquidGlass = canUseLiquidGlass();

    function update() {
      if (!supportsLiquidGlass) {
        layer.classList.add("fallback");
        layer.style.backdropFilter = "";
        layer.style.webkitBackdropFilter = "";
        return;
      }

      layer.classList.remove("fallback");

      const rect = layer.getBoundingClientRect();
      const height = Math.max(1, Math.round(rect.height));
      const width = Math.max(1, Math.round(rect.width));
      const depth = Math.min(
        CONFIG.depth,
        Math.floor(Math.min(height, width) / 3)
      );
      const filter = getDisplacementFilter({
        height,
        width,
        radius: CONFIG.radius,
        depth,
        strength: CONFIG.strength,
        chromaticAberration: CONFIG.chromaticAberration,
      });
      const backdropFilter = `blur(${CONFIG.blur / 2}px) url('${filter}') blur(${CONFIG.blur}px) brightness(1.12) saturate(1.45)`;

      layer.style.backdropFilter = backdropFilter;
      layer.style.webkitBackdropFilter = backdropFilter;
    }

    update();

    if (window.ResizeObserver) {
      const observer = new ResizeObserver(update);
      observer.observe(layer);
    } else {
      window.addEventListener("resize", update);
    }
  }

  function init() {
    document.querySelectorAll(".top-strip").forEach(applyLiquidGlass);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
