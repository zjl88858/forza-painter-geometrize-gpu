package gpu

const evaluateKernelSource = `
__kernel void evaluate_candidates_v2(
    __global const float4* target,
    __global const float4* current,
    __global const uchar* opaqueMask,
    __global const float* candidates,
    __global float* scores,
    const int width,
    const int height,
    const int pixelCount
) {
    int gid = get_global_id(0);

    int base = gid * 9;
    float cx = candidates[base + 0];
    float cy = candidates[base + 1];
    float rx = fmax(candidates[base + 2], 1.0f);
    float ry = fmax(candidates[base + 3], 1.0f);
    float thetaDeg = candidates[base + 4];
    float cr = clamp(candidates[base + 5], 0.0f, 1.0f);
    float cg = clamp(candidates[base + 6], 0.0f, 1.0f);
    float cb = clamp(candidates[base + 7], 0.0f, 1.0f);
    float ca = clamp(candidates[base + 8], 0.0f, 1.0f);

    float theta = thetaDeg * 0.01745329251994329577f;
    float cosT = cos(theta);
    float sinT = sin(theta);
    float invRX2 = 1.0f / (rx * rx);
    float invRY2 = 1.0f / (ry * ry);

    float rx2 = rx * rx;
    float ry2 = ry * ry;
    float cos2 = cosT * cosT;
    float sin2 = sinT * sinT;
    float ex = sqrt(rx2 * cos2 + ry2 * sin2);
    float ey = sqrt(rx2 * sin2 + ry2 * cos2);

    int xMin = (int)floor(cx - ex - 1.0f);
    int xMax = (int)ceil(cx + ex + 1.0f);
    int yMin = (int)floor(cy - ey - 1.0f);
    int yMax = (int)ceil(cy + ey + 1.0f);

    xMin = max(0, xMin);
    yMin = max(0, yMin);
    xMax = min(width - 1, xMax);
    yMax = min(height - 1, yMax);

    float totalDelta = 0.0f;
    int invalidTransparencyOverlap = 0;

    for (int y = yMin; y <= yMax && invalidTransparencyOverlap == 0; ++y) {
        int row = y * width;
        for (int x = xMin; x <= xMax; ++x) {
            int p = row + x;
            float dx = ((float)x + 0.5f) - cx;
            float dy = ((float)y + 0.5f) - cy;
            float xr = dx * cosT + dy * sinT;
            float yr = -dx * sinT + dy * cosT;
            if (xr * xr * invRX2 + yr * yr * invRY2 > 1.0f) {
                continue;
            }

            if (opaqueMask[p] == 0) {
                invalidTransparencyOverlap = 1;
                break;
            }

            float4 src = current[p];
            float4 t = target[p];

            float dr0 = t.x - src.x;
            float dg0 = t.y - src.y;
            float db0 = t.z - src.z;
            float da0 = t.w - src.w;
            float oldErr = dr0 * dr0 + dg0 * dg0 + db0 * db0 + da0 * da0;

            float invA = 1.0f - ca;
            float4 out;
            out.x = src.x * invA + cr * ca;
            out.y = src.y * invA + cg * ca;
            out.z = src.z * invA + cb * ca;
            out.w = src.w * invA + ca;

            float dr1 = t.x - out.x;
            float dg1 = t.y - out.y;
            float db1 = t.z - out.z;
            float da1 = t.w - out.w;
            float newErr = dr1 * dr1 + dg1 * dg1 + db1 * db1 + da1 * da1;

            totalDelta += (newErr - oldErr);
        }
    }

    if (invalidTransparencyOverlap != 0) {
        scores[gid] = 3.402823466e+38f;
    } else {
        scores[gid] = totalDelta;
    }
}

__kernel void apply_candidate_v1(
    __global float4* current,
    __global const uchar* opaqueMask,
    const int width,
    const int height,
    const float cx,
    const float cy,
    const float rxRaw,
    const float ryRaw,
    const float thetaDeg,
    const float cr,
    const float cg,
    const float cb,
    const float ca
) {
    int p = get_global_id(0);
    int pixelCount = width * height;
    if (p >= pixelCount) {
        return;
    }
    if (opaqueMask[p] == 0) {
        return;
    }

    float rx = fmax(rxRaw, 1.0f);
    float ry = fmax(ryRaw, 1.0f);
    float theta = thetaDeg * 0.01745329251994329577f;
    float cosT = cos(theta);
    float sinT = sin(theta);
    float invRX2 = 1.0f / (rx * rx);
    float invRY2 = 1.0f / (ry * ry);

    int x = p % width;
    int y = p / width;
    float dx = ((float)x + 0.5f) - cx;
    float dy = ((float)y + 0.5f) - cy;
    float xr = dx * cosT + dy * sinT;
    float yr = -dx * sinT + dy * cosT;
    if (xr * xr * invRX2 + yr * yr * invRY2 > 1.0f) {
        return;
    }

    float4 src = current[p];
    float alpha = clamp(ca, 0.0f, 1.0f);
    float invA = 1.0f - alpha;
    src.x = src.x * invA + clamp(cr, 0.0f, 1.0f) * alpha;
    src.y = src.y * invA + clamp(cg, 0.0f, 1.0f) * alpha;
    src.z = src.z * invA + clamp(cb, 0.0f, 1.0f) * alpha;
    src.w = src.w * invA + alpha;
    current[p] = src;
}
`
