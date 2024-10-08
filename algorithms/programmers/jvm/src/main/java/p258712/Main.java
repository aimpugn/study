package main.java.p258712;

import java.util.*;

/**
 * @see <a href=
 *      "https://school.programmers.co.kr/learn/courses/30/lessons/258712"> 가장
 *      많이 받은 선물 </a>
 */
public class Main {

    public static void main(String[] args) {
        System.out.println(
                Main.solution(new String[] {"muzi", "ryan", "frodo", "neo"},
                        new String[] {"muzi frodo", "muzi frodo", "ryan muzi",
                                "ryan muzi", "ryan muzi", "frodo muzi",
                                "frodo ryan", "neo muzi"}));
        System.out.println(Main.solution(
                new String[] {"joy", "brad", "alessandro", "conan", "david"},
                new String[] {"alessandro brad", "alessandro joy",
                        "alessandro conan", "david alessandro",
                        "alessandro david"}));
        System.out.println(Main.solution(new String[] {"a", "b", "c"},
                new String[] {"a b", "b a", "c a", "a c", "a c", "c a"}));
    }

    public static int solution(String[] friends, String[] gifts) {
        if (gifts.length == 0)
            return 0;

        Map<String, Integer> friendIndices = createFriendIndices(friends);
        int[][] giftExchanges = new int[friends.length][friends.length];
        int[] giftIndices = new int[friends.length];

        processGifts(gifts, friendIndices, giftExchanges, giftIndices);

        return predictNextMonthGifts(friends, friendIndices, giftExchanges,
                giftIndices);
    }

    private static Map<String, Integer> createFriendIndices(String[] friends) {
        Map<String, Integer> indices = new HashMap<>();
        for (int i = 0; i < friends.length; i++) {
            indices.put(friends[i], i);
        }
        return indices;
    }

    private static void processGifts(String[] gifts,
            Map<String, Integer> friendIndices, int[][] giftExchanges,
            int[] giftIndices) {
        for (String gift : gifts) {
            String[] names = gift.split(" ");
            int giverIndex = friendIndices.get(names[0]);
            int receiverIndex = friendIndices.get(names[1]);

            giftExchanges[giverIndex][receiverIndex]++;
            giftIndices[giverIndex]++;
            giftIndices[receiverIndex]--;
        }
    }

    private static int predictNextMonthGifts(String[] friends,
            Map<String, Integer> friendIndices, int[][] giftExchanges,
            int[] giftIndices) {
        int[] nextMonthGifts = new int[friends.length];

        for (int i = 0; i < friends.length; i++) {
            for (int j = i + 1; j < friends.length; j++) {
                int giftsFromIToJ = giftExchanges[i][j];
                int giftsFromJToI = giftExchanges[j][i];

                if (giftsFromIToJ > giftsFromJToI) {
                    nextMonthGifts[i]++;
                } else if (giftsFromIToJ < giftsFromJToI) {
                    nextMonthGifts[j]++;
                } else if (giftIndices[i] > giftIndices[j]) {
                    nextMonthGifts[i]++;
                } else if (giftIndices[i] < giftIndices[j]) {
                    nextMonthGifts[j]++;
                }
            }
        }

        return Arrays.stream(nextMonthGifts).max().orElse(0);
    }


    /**
     * <pre>
     * 친구들이 이번 달까지 선물을 주고받은 기록을 바탕으로 다음 달에 누가 선물을 많이 받을지 예측.
     *
     * 두 사람이 선물을 주고받은 기록이 있다면, 이번 달까지 두 사람 사이에 더 많은 선물을 준 사람이 다음 달에 선물을 하나 받습니다.
     *  - A -> B에게 선물을 5번
     *  - B -> A에게 선물을 3번
     *  - 다음 달엔 A가 B에게 선물을 하나 받습니다
     *
     * 두 사람이 선물을 주고받은 기록이 하나도 없거나 주고받은 수가 같다면,
     * 선물 지수가 더 큰 사람이 선물 지수가 더 작은 사람에게 선물을 하나 받습니다.
     *  - 선물 지수 = 이번 달까지 자신이 친구들에게 준 선물의 수 - 받은 선물의 수
     *  - A가 친구들에게 준 선물이 3개고 받은 선물이 10개라면 A의 선물 지수는 -7
     *  - B가 친구들에게 준 선물이 3개고 받은 선물이 2개라면 B의 선물 지수는 1입니다.
     *  - A와 B가 선물을 주고받은 적이 없거나 정확히 같은 수로 선물을 주고받았다면,
     *    다음 달엔 B가 A에게 선물을 하나 받습니다.(A -> B)
     *  - 만약 두 사람의 선물 지수도 같다면 다음 달에 선물을 주고받지 않습니다.
     *
     * </pre>
     *
     * @param friends 2 ≤ friends의 길이 = 친구들의 수 ≤ 50. 알파벳 소문자로 이루어진 길이가 10 이하인
     *        문자열. 이름이 같은 친구는 없습니다.
     * @param gifts 1 ≤ gifts의 길이 ≤ 10,000. gifts의 원소는 "A B"형태의 문자열입니다. A는 선물을 준
     *        친구의 이름을 B는 선물을 받은 친구의 이름을 의미하며 공백 하나로 구분됩니다.
     * @return 다음달에 가장 많은 선물을 받는 친구가 받을 선물의 수
     */
    public static int solution_1st(String[] friends, String[] gifts) {
        if (gifts.length == 0) {
            return 0;
        }


        HashMap<String, Integer> relations = new HashMap<>();

        for (String gift : gifts) {
            if (relations.containsKey(gift)) {
                relations.put(gift, relations.get(gift) + 1);
            } else {
                relations.put(gift, 1);
            }
        }

        HashSet<String> checked = new HashSet<>();
        HashMap<String, Integer> indices = new HashMap<>();

        /*
         *
         * A -> B A -> C A -> D
         *
         * B -> A B -> C B -> D
         *
         * A가 선물 지수가 더 높으니까 다음 달에 하나 받는다. 준 경우에는 +1, 받은 경우에는 -1 결국 AB와 BA의 차이를
         * 계산하는 것 아닌가?
         */
        for (Map.Entry<String, Integer> relationEntry : relations.entrySet()) {
            String relation = relationEntry.getKey();

            if (checked.contains(relation)) {
                continue;
            }

            String[] names = relation.split(" ");
            String give = names[0];
            String take = names[1];
            String reverse = take + " " + give;

            int received = 0;

            // "B -> A" 관계
            if (relations.containsKey(reverse)) {
                received = relations.get(reverse);
            }

            // 준 사람과 받은 사람 둘 사이의 지수 계산
            if (indices.containsKey(give)) {
                indices.put(give, indices.get(give) + relationEntry.getValue()
                        - received);
            } else {
                indices.put(give, relationEntry.getValue() - received);
            }

            if (indices.containsKey(take)) {
                indices.put(take, indices.get(take) + received
                        - relationEntry.getValue());
            } else {
                indices.put(take, received - relationEntry.getValue());
            }

            // 역 관계까지 체크했으므로 스킵하도록 합니다.
            checked.add(reverse);
        }

        HashMap<String, Integer> expectation = new HashMap<>();

        for (int i = 0; i < friends.length; i++) {
            String currentName = friends[i];
            for (int j = i + 1; j < friends.length; j++) {
                String other = friends[j];
                String giveRelation = currentName + " " + other;
                String takeRelation = other + " " + currentName;

                int give = 0;
                int take = 0;

                if (relations.containsKey(giveRelation)) {
                    // 선물을 준 관계가 있는 경우
                    give = relations.get(giveRelation);
                }
                if (relations.containsKey(takeRelation)) {
                    // 선물을 받은 관계가 있는 경우
                    take = relations.get(takeRelation);
                }

                String nameWillReceive = "";

                // 주고 받은 수가 다르다면
                // 준 게 더 많다면, 받는다
                if (take < give) {
                    nameWillReceive = currentName;
                } else if (take > give) {
                    nameWillReceive = other;
                } else {
                    // 주고 받은 수가 같거나 관계가 없다면
                    int indexOfCurrentName =
                            indices.getOrDefault(currentName, 0);
                    int indexOfOther = indices.getOrDefault(other, 0);

                    // 선물 지수가 큰 쪽이 받는다
                    if (indexOfOther < indexOfCurrentName) {
                        nameWillReceive = currentName;
                    } else if (indexOfOther > indexOfCurrentName) {
                        nameWillReceive = other;
                    }
                    // 선물 지수도 같다면 아무것도 하지 않는다.
                }

                if (!nameWillReceive.isEmpty()) {
                    if (expectation.containsKey(nameWillReceive)) {
                        expectation.put(nameWillReceive,
                                expectation.get(nameWillReceive) + 1);
                    } else {
                        expectation.put(nameWillReceive, 1);
                    }
                }
            }
        }

        if (expectation.isEmpty()) {
            return 0;
        }


        return Collections.max(expectation.values());
    }
}


class Relation {
    HashMap<String, Integer> friends = new HashMap<>();
}
