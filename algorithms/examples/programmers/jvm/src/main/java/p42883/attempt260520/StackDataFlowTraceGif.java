package p42883.attempt260520;

import javax.imageio.IIOImage;
import javax.imageio.ImageIO;
import javax.imageio.ImageTypeSpecifier;
import javax.imageio.ImageWriteParam;
import javax.imageio.ImageWriter;
import javax.imageio.metadata.IIOMetadata;
import javax.imageio.metadata.IIOMetadataNode;
import javax.imageio.stream.ImageOutputStream;
import java.awt.BasicStroke;
import java.awt.Color;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics2D;
import java.awt.GraphicsEnvironment;
import java.awt.RenderingHints;
import java.awt.geom.RoundRectangle2D;
import java.awt.image.BufferedImage;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Set;

public final class StackDataFlowTraceGif {
    private static final String NUMBER = "4177252841";
    private static final int REMOVE_LIMIT = 4;
    private static final int TARGET_LENGTH = NUMBER.length() - REMOVE_LIMIT;
    private static final Path OUTPUT = Path.of(
        "src/main/java/p42883/attempt260520/stack_data_flow_trace.gif"
    );

    private static final int WIDTH = 1600;
    private static final int HEIGHT = 1000;
    private static final int IMAGE_TYPE = BufferedImage.TYPE_INT_RGB;

    private static final Color BACKGROUND = new Color(246, 248, 251);
    private static final Color INK = new Color(35, 42, 52);
    private static final Color MUTED = new Color(98, 110, 125);
    private static final Color CARD = new Color(255, 255, 255);
    private static final Color LINE = new Color(218, 225, 234);
    private static final Color BLUE = new Color(41, 99, 255);
    private static final Color BLUE_SOFT = new Color(226, 236, 255);
    private static final Color GREEN = new Color(23, 139, 89);
    private static final Color GREEN_SOFT = new Color(224, 244, 234);
    private static final Color RED = new Color(217, 72, 80);
    private static final Color RED_SOFT = new Color(255, 230, 232);
    private static final Color AMBER = new Color(245, 169, 35);
    private static final Color AMBER_SOFT = new Color(255, 245, 216);
    private static final Color GRAY_SOFT = new Color(238, 242, 247);
    private static final Color DARK_PANEL = new Color(42, 50, 63);
    private static final String FONT_FAMILY = chooseFontFamily();

    private StackDataFlowTraceGif() {
    }

    public static void main(String[] args) throws IOException {
        var frames = buildFrames();
        Files.createDirectories(OUTPUT.getParent());
        writeGif(frames, OUTPUT);
        System.out.println("wrote " + OUTPUT.toAbsolutePath());
    }

    private static List<FrameState> buildFrames() {
        var frames = new ArrayList<FrameState>();
        var stack = new ArrayList<Token>();
        var removed = new LinkedHashSet<Integer>();
        var history = new ArrayList<TraceRow>();
        var removeLeft = REMOVE_LIMIT;

        frames.add(new FrameState(
            -1,
            "시작 상태",
            "정답 길이 " + TARGET_LENGTH + "을 만들기 위해 " + REMOVE_LIMIT + "개를 지울 수 있습니다.",
            "규칙: 새 숫자가 stack 끝보다 크면, 지울 수 있는 동안 작은 끝 숫자부터 제거합니다.",
            List.copyOf(stack),
            new LinkedHashSet<>(removed),
            removeLeft,
            List.copyOf(history),
            false,
            2600
        ));

        for (int i = 0; i < NUMBER.length(); i++) {
            var digit = NUMBER.charAt(i);
            var popped = false;

            while (removeLeft > 0 && !stack.isEmpty() && stack.getLast().digit() < digit) {
                var removedToken = stack.removeLast();
                removeLeft--;
                removed.add(removedToken.index());
                popped = true;
                history.add(new TraceRow(
                    "pop " + removedToken.digit(),
                    removedToken.digit() + " < " + digit,
                    stackText(stack),
                    removeLeft
                ));
                frames.add(new FrameState(
                    i,
                    "pop: " + removedToken.digit() + " 제거",
                    removedToken.digit() + "은 뒤에서 온 " + digit + "보다 작아서 앞자리를 약하게 만듭니다.",
                    "삭제 예산이 남아 있으므로 stack 끝의 작은 숫자를 먼저 지웁니다.",
                    List.copyOf(stack),
                    new LinkedHashSet<>(removed),
                    removeLeft,
                    recent(history),
                    false,
                    1700
                ));
            }

            stack.add(new Token(i, digit));
            history.add(new TraceRow(
                "push " + digit,
                popped ? "더 지울 수 없거나 top >= " + digit : "top이 없거나 top >= " + digit,
                stackText(stack),
                removeLeft
            ));
            frames.add(new FrameState(
                i,
                "push: " + digit + " 보관",
                popped
                    ? "작은 끝 숫자를 정리한 뒤 현재 숫자를 뒤에 붙입니다."
                    : "현재 숫자가 앞 숫자를 밀어낼 조건이 아니므로 그대로 보관합니다.",
                "stack은 지금까지 남기기로 한 숫자를 왼쪽에서 오른쪽 순서로 저장합니다.",
                List.copyOf(stack),
                new LinkedHashSet<>(removed),
                removeLeft,
                recent(history),
                false,
                1500
            ));
        }

        frames.add(new FrameState(
            NUMBER.length() - 1,
            "완성: " + stackText(stack),
            "삭제 " + REMOVE_LIMIT + "개를 모두 썼고, 남은 숫자를 이어 붙이면 " + stackText(stack) + "입니다.",
            "핵심 기억: 더 큰 숫자가 뒤에서 오면, 직전에 남긴 작은 숫자부터 지웁니다.",
            List.copyOf(stack),
            new LinkedHashSet<>(removed),
            0,
            recent(history),
            true,
            2800
        ));

        return frames;
    }

    private static List<TraceRow> recent(List<TraceRow> history) {
        return List.copyOf(history.subList(Math.max(0, history.size() - 6), history.size()));
    }

    private static String stackText(List<Token> stack) {
        var value = new StringBuilder(stack.size());
        for (var token : stack) {
            value.append(token.digit());
        }
        return value.toString();
    }

    private static void writeGif(List<FrameState> frames, Path output) throws IOException {
        var writers = ImageIO.getImageWritersBySuffix("gif");
        if (!writers.hasNext()) {
            throw new IllegalStateException("No GIF writer is available");
        }

        var writer = writers.next();
        var param = writer.getDefaultWriteParam();
        try (ImageOutputStream stream = ImageIO.createImageOutputStream(output.toFile())) {
            writer.setOutput(stream);
            writer.prepareWriteSequence(null);
            for (int i = 0; i < frames.size(); i++) {
                var image = drawFrame(frames.get(i), i + 1, frames.size());
                var metadata = metadata(writer, param, frames.get(i).durationMs(), i == 0);
                writer.writeToSequence(new IIOImage(image, null, metadata), param);
            }
            writer.endWriteSequence();
        } finally {
            writer.dispose();
        }
    }

    private static IIOMetadata metadata(
        ImageWriter writer,
        ImageWriteParam param,
        int delayMs,
        boolean firstFrame
    ) throws IOException {
        var imageType = ImageTypeSpecifier.createFromBufferedImageType(IMAGE_TYPE);
        var metadata = writer.getDefaultImageMetadata(imageType, param);
        var format = metadata.getNativeMetadataFormatName();
        var root = (IIOMetadataNode) metadata.getAsTree(format);

        var gce = node(root, "GraphicControlExtension");
        gce.setAttribute("disposalMethod", "none");
        gce.setAttribute("userInputFlag", "FALSE");
        gce.setAttribute("transparentColorFlag", "FALSE");
        gce.setAttribute("delayTime", Integer.toString(Math.max(2, Math.round(delayMs / 10.0f))));
        gce.setAttribute("transparentColorIndex", "0");

        if (firstFrame) {
            var appExtensions = node(root, "ApplicationExtensions");
            var app = new IIOMetadataNode("ApplicationExtension");
            app.setAttribute("applicationID", "NETSCAPE");
            app.setAttribute("authenticationCode", "2.0");
            app.setUserObject(new byte[] {0x1, 0x0, 0x0});
            appExtensions.appendChild(app);
        }

        metadata.setFromTree(format, root);
        return metadata;
    }

    private static IIOMetadataNode node(IIOMetadataNode root, String name) {
        var nodes = root.getElementsByTagName(name);
        if (nodes.getLength() > 0) {
            return (IIOMetadataNode) nodes.item(0);
        }
        var node = new IIOMetadataNode(name);
        root.appendChild(node);
        return node;
    }

    private static BufferedImage drawFrame(FrameState state, int frameNumber, int totalFrames) {
        var image = new BufferedImage(WIDTH, HEIGHT, IMAGE_TYPE);
        var g = image.createGraphics();
        try {
            setup(g);
            g.setColor(BACKGROUND);
            g.fillRect(0, 0, WIDTH, HEIGHT);

            drawHeader(g, state, frameNumber, totalFrames);
            drawInputStrip(g, state);
            drawRulePanel(g, state);
            drawStack(g, state);
            drawBudget(g, state);
            drawTrace(g, state);
            if (state.complete()) {
                drawResultRibbon(g, state);
            }
        } finally {
            g.dispose();
        }
        return image;
    }

    private static void setup(Graphics2D g) {
        g.setRenderingHint(RenderingHints.KEY_ANTIALIASING, RenderingHints.VALUE_ANTIALIAS_ON);
        g.setRenderingHint(RenderingHints.KEY_TEXT_ANTIALIASING, RenderingHints.VALUE_TEXT_ANTIALIAS_ON);
        g.setRenderingHint(RenderingHints.KEY_RENDERING, RenderingHints.VALUE_RENDER_QUALITY);
    }

    private static void drawHeader(Graphics2D g, FrameState state, int frameNumber, int totalFrames) {
        text(g, "큰 수 만들기: stack으로 앞자리를 키우는 과정", 70, 78, 36, Font.BOLD, INK);
        text(g, "입력 " + NUMBER + ", k=" + REMOVE_LIMIT + ", 정답 길이=" + TARGET_LENGTH, 72, 118, 23, Font.PLAIN, MUTED);

        var progress = frameNumber + " / " + totalFrames;
        var metrics = metrics(g, 20, Font.BOLD);
        var width = metrics.stringWidth(progress) + 42;
        round(g, WIDTH - 72 - width, 54, width, 44, 18, DARK_PANEL, null, 1);
        text(g, progress, WIDTH - 72 - width + 22, 83, 20, Font.BOLD, Color.WHITE);
    }

    private static void drawInputStrip(Graphics2D g, FrameState state) {
        round(g, 70, 155, 1460, 160, 28, CARD, LINE, 1);
        text(g, "입력 숫자", 100, 202, 24, Font.BOLD, INK);
        text(g, "노랑=현재, 붉은 X=삭제됨", 100, 236, 18, Font.PLAIN, MUTED);

        var startX = 485;
        var y = 190;
        var size = 78;
        var gap = 16;
        for (int i = 0; i < NUMBER.length(); i++) {
            var fill = GRAY_SOFT;
            var stroke = LINE;
            if (state.removed().contains(i)) {
                fill = RED_SOFT;
                stroke = RED;
            } else if (i < state.inputIndex() || state.complete()) {
                fill = GREEN_SOFT;
                stroke = GREEN;
            } else if (i == state.inputIndex()) {
                fill = AMBER_SOFT;
                stroke = AMBER;
            }

            var x = startX + i * (size + gap);
            round(g, x, y, size, size, 18, fill, stroke, 2);
            centered(g, String.valueOf(NUMBER.charAt(i)), x, y + 6, size, size, 34, Font.BOLD, INK);
            centered(g, String.valueOf(i), x, y + size + 8, size, 24, 15, Font.PLAIN, MUTED);

            if (state.removed().contains(i)) {
                g.setColor(RED);
                g.setStroke(new BasicStroke(5, BasicStroke.CAP_ROUND, BasicStroke.JOIN_ROUND));
                g.drawLine(x + 18, y + 18, x + size - 18, y + size - 18);
                g.drawLine(x + size - 18, y + 18, x + 18, y + size - 18);
            }
        }
    }

    private static void drawRulePanel(Graphics2D g, FrameState state) {
        round(g, 70, 345, 620, 275, 28, CARD, LINE, 1);
        text(g, "이번 판단", 105, 393, 26, Font.BOLD, INK);
        text(g, state.title(), 105, 444, 34, Font.BOLD, state.complete() ? GREEN : BLUE);
        wrappedText(g, state.description(), 105, 485, 540, 24, 23, INK);

        round(g, 105, 548, 520, 46, 16, BLUE_SOFT, null, 1);
        text(g, state.rule(), 125, 579, 18, Font.BOLD, new Color(31, 70, 160));
    }

    private static void drawStack(Graphics2D g, FrameState state) {
        round(g, 730, 345, 800, 275, 28, CARD, LINE, 1);
        text(g, "stack: 지금까지 남기기로 한 숫자", 765, 393, 26, Font.BOLD, INK);
        text(g, "오른쪽 끝이 다음 비교 대상입니다.", 765, 424, 18, Font.PLAIN, MUTED);

        var box = 76;
        var gap = 14;
        var totalWidth = state.stack().size() * box + Math.max(0, state.stack().size() - 1) * gap;
        var x = 765 + Math.max(0, (710 - totalWidth) / 2);
        var y = 478;
        if (state.stack().isEmpty()) {
            centered(g, "비어 있음", 765, y, 710, 74, 28, Font.BOLD, MUTED);
        }
        for (int i = 0; i < state.stack().size(); i++) {
            var token = state.stack().get(i);
            var isTop = i == state.stack().size() - 1;
            round(g, x, y, box, box, 18, isTop ? AMBER_SOFT : BLUE_SOFT, isTop ? AMBER : BLUE, 2);
            centered(g, String.valueOf(token.digit()), x, y + 4, box, box, 34, Font.BOLD, INK);
            centered(g, "idx " + token.index(), x, y + box + 7, box, 20, 14, Font.PLAIN, MUTED);
            x += box + gap;
        }
    }

    private static void drawBudget(Graphics2D g, FrameState state) {
        round(g, 70, 650, 620, 130, 28, CARD, LINE, 1);
        text(g, "삭제 예산", 105, 700, 25, Font.BOLD, INK);
        text(g, "남은 삭제 횟수: " + state.removeLeft(), 105, 733, 20, Font.PLAIN, MUTED);

        var used = REMOVE_LIMIT - state.removeLeft();
        for (int i = 0; i < REMOVE_LIMIT; i++) {
            var x = 370 + i * 62;
            var fill = i < used ? RED_SOFT : GREEN_SOFT;
            var stroke = i < used ? RED : GREEN;
            round(g, x, 685, 48, 48, 14, fill, stroke, 2);
            centered(g, i < used ? "X" : "+", x, 687, 48, 48, 24, Font.BOLD, stroke);
        }
    }

    private static void drawTrace(Graphics2D g, FrameState state) {
        round(g, 730, 650, 800, 275, 28, CARD, LINE, 1);
        text(g, "최근 trace", 765, 698, 25, Font.BOLD, INK);
        text(g, "action -> 이유 -> stack -> 남은 삭제", 765, 728, 18, Font.PLAIN, MUTED);

        var y = 765;
        for (var row : state.history()) {
            var actionColor = row.action().startsWith("pop") ? RED : GREEN;
            text(g, row.action(), 765, y, 18, Font.BOLD, actionColor);
            text(g, row.reason(), 875, y, 17, Font.PLAIN, INK);
            text(g, "[" + row.stackAfter() + "]", 1145, y, 17, Font.BOLD, BLUE);
            text(g, "rem=" + row.removeLeft(), 1355, y, 17, Font.PLAIN, MUTED);
            y += 31;
        }
    }

    private static void drawResultRibbon(Graphics2D g, FrameState state) {
        round(g, 70, 810, 620, 115, 28, new Color(31, 130, 87), null, 1);
        text(g, "정답", 110, 862, 26, Font.BOLD, Color.WHITE);
        text(g, stackText(state.stack()), 220, 879, 54, Font.BOLD, Color.WHITE);
        text(g, "앞자리부터 가장 크게 남긴 결과입니다.", 110, 905, 19, Font.PLAIN, new Color(228, 255, 242));
    }

    private static void round(
        Graphics2D g,
        int x,
        int y,
        int width,
        int height,
        int radius,
        Color fill,
        Color stroke,
        int strokeWidth
    ) {
        var shape = new RoundRectangle2D.Double(x, y, width, height, radius, radius);
        if (fill != null) {
            g.setColor(fill);
            g.fill(shape);
        }
        if (stroke != null) {
            g.setColor(stroke);
            g.setStroke(new BasicStroke(strokeWidth));
            g.draw(shape);
        }
    }

    private static void text(Graphics2D g, String value, int x, int y, int size, int style, Color color) {
        g.setFont(font(size, style));
        g.setColor(color);
        g.drawString(value, x, y);
    }

    private static void wrappedText(
        Graphics2D g,
        String value,
        int x,
        int y,
        int maxWidth,
        int lineHeight,
        int size,
        Color color
    ) {
        g.setFont(font(size, Font.PLAIN));
        g.setColor(color);
        var metrics = g.getFontMetrics();
        var line = new StringBuilder();
        var currentY = y;
        for (var word : value.split(" ")) {
            var candidate = line.isEmpty() ? word : line + " " + word;
            if (metrics.stringWidth(candidate) > maxWidth && !line.isEmpty()) {
                g.drawString(line.toString(), x, currentY);
                line.setLength(0);
                line.append(word);
                currentY += lineHeight;
            } else {
                line.setLength(0);
                line.append(candidate);
            }
        }
        if (!line.isEmpty()) {
            g.drawString(line.toString(), x, currentY);
        }
    }

    private static void centered(
        Graphics2D g,
        String value,
        int x,
        int y,
        int width,
        int height,
        int size,
        int style,
        Color color
    ) {
        g.setFont(font(size, style));
        g.setColor(color);
        var metrics = g.getFontMetrics();
        var textX = x + (width - metrics.stringWidth(value)) / 2;
        var textY = y + (height - metrics.getHeight()) / 2 + metrics.getAscent();
        g.drawString(value, textX, textY);
    }

    private static FontMetrics metrics(Graphics2D g, int size, int style) {
        g.setFont(font(size, style));
        return g.getFontMetrics();
    }

    private static Font font(int size, int style) {
        return new Font(FONT_FAMILY, style, size);
    }

    private static String chooseFontFamily() {
        var preferred = "Apple SD Gothic Neo";
        var available = Set.of(GraphicsEnvironment.getLocalGraphicsEnvironment().getAvailableFontFamilyNames());
        return available.contains(preferred) ? preferred : Font.SANS_SERIF;
    }

    private record Token(int index, char digit) {
    }

    private record TraceRow(String action, String reason, String stackAfter, int removeLeft) {
    }

    private record FrameState(
        int inputIndex,
        String title,
        String description,
        String rule,
        List<Token> stack,
        Set<Integer> removed,
        int removeLeft,
        List<TraceRow> history,
        boolean complete,
        int durationMs
    ) {
    }
}
