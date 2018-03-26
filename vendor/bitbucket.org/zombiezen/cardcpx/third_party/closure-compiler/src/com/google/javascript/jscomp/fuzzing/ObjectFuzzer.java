/*
 * Copyright 2013 The Closure Compiler Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package com.google.javascript.jscomp.fuzzing;

import com.google.common.collect.Sets;
import com.google.javascript.rhino.Node;
import com.google.javascript.rhino.Token;

import java.util.Arrays;
import java.util.Set;

/**
 * UNDER DEVELOPMENT. DO NOT USE!
 */
class ObjectFuzzer extends AbstractFuzzer {

  ObjectFuzzer(FuzzingContext context) {
    super(context);
  }

  /* (non-Javadoc)
   * @see com.google.javascript.jscomp.fuzzing.AbstractFuzzer#generate(int)
   */
  @Override
  protected Node generate(int budget, Set<Type> types) {
    Node objectLit = new Node(Token.OBJECTLIT);
    int remainingBudget = budget - 1;
    if (remainingBudget < 0) {
      remainingBudget = 0;
    }
    // an object property needs at least two nodes
    int objectLength = generateLength(remainingBudget / 2);
    if (objectLength == 0) {
      return objectLit;
    }
    // reserve budget for keys
    remainingBudget -= objectLength;
    ExpressionFuzzer[] fuzzers = new ExpressionFuzzer[objectLength];
    Arrays.fill(fuzzers,
        new ExpressionFuzzer(context));
    Node[] values = distribute(remainingBudget, fuzzers);
    for (int i = 0; i < objectLength; i++) {
      String name;
      if (context.random.nextInt(2) == 0) {
        name = context.snGenerator.getPropertyName();
      } else {
        name = String.valueOf(context.snGenerator.getRandomNumber());
      }
      Node key = Node.newString(Token.STRING_KEY, name);

      key.addChildrenToFront(values[i]);
      objectLit.addChildToBack(key);
    }
    return objectLit;
  }

  /* (non-Javadoc)
   * @see com.google.javascript.jscomp.fuzzing.AbstractFuzzer#isEnough(int)
   */
  @Override
  protected boolean isEnough(int budget) {
    return budget >= 1;
  }

  /* (non-Javadoc)
   * @see com.google.javascript.jscomp.fuzzing.AbstractFuzzer#getConfigName()
   */
  @Override
  protected String getConfigName() {
    return "object";
  }

  @Override
  protected Set<Type> supportedTypes() {
    return Sets.immutableEnumSet(Type.OBJECT);
  }
}
